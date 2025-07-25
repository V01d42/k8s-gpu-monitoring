# Multi-stage build for React frontend
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Install all dependencies (including devDependencies for build)
RUN npm ci

# Expose port
EXPOSE 5173

# Start Vite development server
CMD ["npm", "run", "dev"]