docker run --name mongodb -p 27017:27017 -d mongodb/mongodb-community-server:latest

export MONGODB_URI="mongodb://localhost:27017"
