FROM node:18

# Set the working directory
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to the working directory
COPY ./package*.json ./

# Copy the rest of the application code to the working directory
COPY . .

COPY ./certificates/ ./certificates/

RUN npm install
# Expose the ports for the OIDC provider and the Trust Score API
EXPOSE 8080

# Start both the OIDC provider and API server using PM2
CMD ["node", "index.mjs"]