import serverlessExpress from '@codegenie/serverless-express';
import app from './index.mjs';  // Import the Express app

// Create the serverless express instance
const serverlessExpressInstance = serverlessExpress({
  app,
  // log (uncomment if you want logging)
});

// AWS Lambda handler function
export async function handler(event, context) {
  // Pass the event and context to the serverless express instance
  return serverlessExpressInstance(event, context);
}