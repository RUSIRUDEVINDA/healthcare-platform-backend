package provider

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"healthcare-platform/services/payment-service/internal/model"
)

type PayHereProvider struct {
	MerchantID     string
	MerchantSecret string
	AppID          string
	AppSecret      string
	Env            string // "sandbox" or "live"
	ReturnURL      string
	CancelURL      string
	NotifyURL      string
	httpClient     *http.Client
}

func NewPayHereProvider(merchantID, merchantSecret, appID, appSecret, env, returnURL, cancelURL, notifyURL string) *PayHereProvider {
	return &PayHereProvider{
		MerchantID:     merchantID,
		MerchantSecret: merchantSecret,
		AppID:          appID,
		AppSecret:      appSecret,
		Env:            env,
		ReturnURL:      returnURL,
		CancelURL:      cancelURL,
		NotifyURL:      notifyURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (p *PayHereProvider) Name() string {
	return "payhere"
}

func (p *PayHereProvider) CheckoutURL() string {
	if strings.EqualFold(p.Env, "live") || strings.EqualFold(p.Env, "production") {
		return "https://www.payhere.lk/pay/checkout"
	}
	return "https://sandbox.payhere.lk/pay/checkout"
}

func (p *PayHereProvider) apiBaseURL() string {
	if strings.EqualFold(p.Env, "live") || strings.EqualFold(p.Env, "production") {
		return "https://www.payhere.lk/merchant/v1"
	}
	return "https://sandbox.payhere.lk/merchant/v1"
}

func (p *PayHereProvider) BuildCheckout(payment *model.Payment, req *model.CheckoutRequest) (*model.CheckoutResponse, error) {
	if payment == nil {
		return nil, fmt.Errorf("payhere.BuildCheckout: payment is nil")
	}
	if strings.TrimSpace(p.MerchantID) == "" || strings.TrimSpace(p.MerchantSecret) == "" {
		return nil, fmt.Errorf("payhere.BuildCheckout: PAYHERE_MERCHANT_ID/PAYHERE_MERCHANT_SECRET not configured")
	}
	if strings.TrimSpace(p.ReturnURL) == "" || strings.TrimSpace(p.CancelURL) == "" || strings.TrimSpace(p.NotifyURL) == "" {
		return nil, fmt.Errorf("payhere.BuildCheckout: PAYHERE_RETURN_URL/PAYHERE_CANCEL_URL/PAYHERE_NOTIFY_URL not configured")
	}
	
	orderID := payment.ID
	amount := fmt.Sprintf("%.2f", payment.Amount)
	hash := p.checkoutHash(orderID, amount, payment.Currency)

	items := strings.TrimSpace(req.Items)
	if items == "" {
		items = "Appointment Fee"
	}

	fields := map[string]string{
		"merchant_id": p.MerchantID,
		"return_url":  p.ReturnURL,
		"cancel_url":  p.CancelURL,
		"notify_url":  p.NotifyURL,

		"order_id": orderID,
		"items":    items,
		"currency": payment.Currency,
		"amount":   amount,

		"first_name": req.Customer.FirstName,
		"last_name":  req.Customer.LastName,
		"email":      req.Customer.Email,
		"phone":      req.Customer.Phone,
		"address":    req.Customer.Address,
		"city":       req.Customer.City,
		"country":    req.Customer.Country,

		"custom_1": payment.AppointmentID,
		"custom_2": payment.PatientID,

		"hash": hash,
	}

	return &model.CheckoutResponse{
		PaymentID:   payment.ID,
		Provider:    p.Name(),
		CheckoutURL: p.CheckoutURL(),
		Fields:      fields,
	}, nil
}

func (p *PayHereProvider) VerifyNotification(n *model.PayHereNotification) (bool, error) {
	if n == nil {
		return false, fmt.Errorf("payhere.VerifyNotification: notification is nil")
	}
	if !strings.EqualFold(strings.TrimSpace(n.MerchantID), strings.TrimSpace(p.MerchantID)) {
		return false, nil
	}
	if strings.TrimSpace(n.OrderID) == "" || strings.TrimSpace(n.MD5Sig) == "" {
		return false, nil
	}
	expected := p.notificationSig(n.OrderID, n.PayHereAmount, n.PayHereCurrency, n.StatusCode)
	return strings.EqualFold(expected, n.MD5Sig), nil
}

func (p *PayHereProvider) MapStatus(statusCode int) model.PaymentStatus {
	// PayHere statuses: 2=success, 0=pending, -1=canceled, -2=failed
	switch statusCode {
	case 2:
		return model.StatusCompleted
	case 0:
		return model.StatusPending
	case 3: // Refunded
		return model.StatusRefunded
	default:
		return model.StatusFailed
	}
}

func (p *PayHereProvider) Refund(paymentID string, amount float64) error {
	if p.AppID == "" || p.AppSecret == "" {
		return fmt.Errorf("payhere.Refund: PAYHERE_APP_ID/PAYHERE_APP_SECRET not configured")
	}

	token, err := p.getAccessToken()
	if err != nil {
		return fmt.Errorf("payhere.Refund get token: %w", err)
	}

	refundReq := map[string]interface{}{
		"payment_id":  paymentID,
		"description": "Doctor cancelled appointment",
		"amount":       amount, // Use float64 as PayHere prefers numbers for amounts in JSON
	}
	
	body, _ := json.Marshal(refundReq)
	req, _ := http.NewRequest(http.MethodPost, p.apiBaseURL()+"/payment/refund", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("payhere.Refund request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payhere.Refund failed with HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Status  int    `json:"status"`
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("payhere.Refund decode response: %w (body: %s)", err, string(respBody))
	}

	if result.Status != 1 {
		return fmt.Errorf("payhere.Refund API error (status %d): %s", result.Status, result.Message)
	}

	return nil
}

func (p *PayHereProvider) getAccessToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	auth := base64.StdEncoding.EncodeToString([]byte(p.AppID + ":" + p.AppSecret))
	
	req, _ := http.NewRequest(http.MethodPost, p.apiBaseURL()+"/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func (p *PayHereProvider) checkoutHash(orderID, amount, currency string) string {
	secretHash := strings.ToUpper(md5Hex(p.MerchantSecret))
	raw := p.MerchantID + orderID + amount + currency + secretHash
	return strings.ToUpper(md5Hex(raw))
}

func (p *PayHereProvider) notificationSig(orderID, payhereAmount, payhereCurrency string, statusCode int) string {
	secret := strings.ToUpper(md5Hex(p.MerchantSecret))
	raw := p.MerchantID + orderID + payhereAmount + payhereCurrency + strconv.Itoa(statusCode) + secret
	return strings.ToUpper(md5Hex(raw))
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
