# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install pnpm and dependencies
COPY package.json pnpm-lock.yaml* ./
RUN npm install -g pnpm && pnpm install

# Copy source code and build Vue app
COPY . .
RUN pnpm build-only

# Production stage
FROM nginx:1.27-alpine

# Copy built assets to Nginx html directory
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
