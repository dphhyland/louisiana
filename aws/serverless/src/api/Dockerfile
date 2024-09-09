# Use an official Node.js runtime as a parent image
FROM node:18

# Set the working directory
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to the working directory
COPY ./package*.json ./

# Install any needed dependencies
RUN npm install

# Copy the rest of the application code to the working directory
COPY . .

COPY ./certificates/ ./certificates/

# Expose the ports for the OIDC provider and the Trust Score API
EXPOSE 3000 4000

# Use a process manager to run both services concurrently (like PM2 or npm-run-all)
RUN npm install -g pm2

# Start both the OIDC provider and API server using PM2
CMD ["pm2-runtime", "src/index.mjs"]