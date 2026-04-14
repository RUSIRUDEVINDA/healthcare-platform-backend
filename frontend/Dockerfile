# Stage 1: Build the React Application
FROM node:20-alpine AS build
WORKDIR /app

# Ensure devDependencies are installed
ENV NODE_ENV=development

COPY package*.json ./
RUN npm install

# Explicitly add binaries to path
ENV PATH /app/node_modules/.bin:$PATH

COPY . .
RUN npm run build

# Stage 2: Serve using Nginx
FROM nginx:1.25-alpine
# Copy the built assets
COPY --from=build /app/dist /usr/share/nginx/html
# Copy custom Nginx configuration for the React App
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
