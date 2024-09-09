import ClientsMongoAdapter from './clientsMongoAdapter.js'; // Adjust the path as needed
import MongoAdapter from './mongodb.js'; // Adjust the path as needed

/**
 * Factory function to return the appropriate MongoDB adapter for 'Client' entities.
 *
 * @param {*} DB - The MongoDB instance
 * @returns {Function} A function that returns the correct adapter for a given entity name.
 */
export default function getMongoDBAdapterFactory(DB) {
  return function MongoDBAdapterFactory(name) {
    // Return a ClientsMongoAdapter for 'Client' entities
    if (name === 'Client') {
      return new ClientsMongoAdapter(name, DB);
    }

    // Default to the base MongoAdapter for any other entities
    return new MongoAdapter(name, DB);
  };
}