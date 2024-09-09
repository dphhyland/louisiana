// initializeAdapter.js

import MongoAdapter from './mongodb.js'; // Make sure to add .js extension
import getMongoDBAdapterFactory from './getMongoDBAdapterFactory.js'; // Make sure to add .js extension

/**
 * Initialize MongoDB adapter and return the adapter factory.
 *
 * @returns {Promise<Function>} A function that returns the correct adapter for a given entity name.
 */
async function initializeAdapter() {
  // Step 1: Connect to the MongoDB database
  await MongoAdapter.connect();

  // Step 2: Get the adapter factory
  const adapterFactory = getMongoDBAdapterFactory();

  return adapterFactory;
}

// Export the function as the default export
export default initializeAdapter;