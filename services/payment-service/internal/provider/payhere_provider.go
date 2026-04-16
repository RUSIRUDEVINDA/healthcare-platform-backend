package provider

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"healthcare-platform/services/payment-service/internal/model"
)

type PayHereProvider struct {
	MerchantID     string
	MerchantSecret string
	Env            string // "sandbox" or "live"
	ReturnURL      string
	CancelURL      string
	NotifyURL      string
}

func NewPayHereProvider(merchantID, merchantSecret, env, returnURL, cancelURL, notifyURL string) *PayHereProvider {
	return &PayHereProvider{
		MerchantID:     merchantID,
		MerchantSecret: merchantSecret,
		Env:            env,
		ReturnURL:      returnURL,
		CancelURL:      cancelURL,
		NotifyURL:      notifyURL,
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
	if strings.Contains(p.NotifyURL, "localhost") || strings.Contains(p.NotifyURL, "127.0.0.1") {
		return nil, fmt.Errorf("payhere.BuildCheckout: PAYHERE_NOTIFY_URL must be publicly reachable (use ngrok), got %q", p.NotifyURL)
	}
	if strings.TrimSpace(req.Customer.FirstName) == "" ||
		strings.TrimSpace(req.Customer.LastName) == "" ||
		strings.TrimSpace(req.Customer.Email) == "" ||
		strings.TrimSpace(req.Customer.Phone) == "" ||
		strings.TrimSpace(req.Customer.Address) == "" ||
		strings.TrimSpace(req.Customer.City) == "" ||
		strings.TrimSpace(req.Customer.Country) == "" {
		return nil, fmt.Errorf("payhere.BuildCheckout: missing required customer fields")
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
	// PayHere docs/community: 2=success, 0=pending, -1=canceled, -2=failed, -3=charged-back
	if statusCode == 2 {
		return model.StatusCompleted
	}
	if statusCode == 0 {
		return model.StatusPending
	}
	return model.StatusFailed
}

func (p *PayHereProvider) checkoutHash(orderID, amount, currency string) string {
	merchantIdClean := strings.TrimSpace(p.MerchantID)
	secretClean := strings.TrimSpace(p.MerchantSecret)
	
	secretHash := strings.ToUpper(md5Hex(secretClean))
	raw := merchantIdClean + orderID + amount + currency + secretHash
	
	fmt.Printf("[PAYHERE DEBUG] Generating Hash. MerchantID: '%s', OrderID: '%s', Amount: '%s', Currency: '%s', SecretHash: '%s' \n", 
		merchantIdClean, orderID, amount, currency, secretHash)
	
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
