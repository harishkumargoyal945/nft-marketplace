FROM node:20-alpine

WORKDIR /app

# Install dependencies first for caching
COPY package.json package-lock.json* ./
RUN npm install --save-dev hardhat @nomicfoundation/hardhat-toolbox @openzeppelin/contracts

# Copy the rest
COPY . .

# Expose JSON-RPC port
EXPOSE 8545

# Default command to run node
CMD ["npx", "hardhat", "node"]
