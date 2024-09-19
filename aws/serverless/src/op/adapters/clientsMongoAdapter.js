import axios from 'axios';
import jwt from 'jsonwebtoken';
import MongoAdapter from './mongodb.js'; // Adjust the path as necessary

const FIVE_MINUTES_IN_MS = 5 * 60 * 1000; // 5 minutes in milliseconds

/**
 * Helper function to filter out null values.
 *
 * @param {Object} obj - The object to be filtered.
 * @returns {Object} - A new object with only valid (non-null) key-value pairs.
 */
function filterNullFields(obj) {
  return Object.fromEntries(
    Object.entries(obj)
      .filter(([_, value]) => value !== null) // Filter out only null values
  );
}

class ClientsMongoAdapter extends MongoAdapter {
  constructor(name) {
    super(name);
  }

  /**
   * Override the find method to implement find-or-fetch-create logic, including refresh logic.
   *
   * @param {string} _id - The client_id (URI)
   * @returns {Object | undefined} The client payload
   */
  async find(_id) {
    // Step 1: Look up the client in MongoDB
    const existingClient = await super.find(_id);

    // Check if the client exists and needs to be refreshed
    if (existingClient && this.needsRefresh(existingClient.federationCreationTimestamp)) {
      console.log(`Refreshing client ${_id} metadata from federation.`);
      return await this.refreshClient(_id, existingClient);
    }

    if (existingClient) {
      return existingClient;
    }

    // Step 2: Fetch new client metadata if it does not exist
    return await this.createNewClient(_id);
  }

  /**
   * Check if the client needs to be refreshed based on federationCreationTimestamp.
   *
   * @param {string} federationCreationTimestamp - The timestamp when the client was last fetched.
   * @returns {boolean} - Returns true if the client needs to be refreshed.
   */
  needsRefresh(federationCreationTimestamp) {
    const now = new Date().getTime();
    const lastUpdated = new Date(federationCreationTimestamp).getTime();
    return now - lastUpdated > FIVE_MINUTES_IN_MS;
  }

  /**
   * Refresh the client metadata from the federation and upsert it into the database.
   *
   * @param {string} _id - The client_id (URI)
   * @param {Object} existingClient - The existing client object
   * @returns {Object | undefined} The refreshed client object
   */
  async refreshClient(_id, existingClient) {
    // Fetch new metadata from the federation endpoint
    const newClient = await this.fetchClientMetadata(_id);

    if (!newClient) {
      console.error(`Failed to refresh client ${_id} from federation.`);
      return existingClient; // Return existing client if refresh fails
    }

    // Update federationCreationTimestamp to the current time
    newClient.federationCreationTimestamp = new Date().toISOString();

    // Upsert the refreshed client into the database
    try {
      await this.upsert(_id, newClient);
    } catch (error) {
      console.error(`Error updating client ${_id} in DB:`, error);
      return existingClient; // Return the existing client if the update fails
    }

    return newClient;
  }

  /**
   * Fetch new client metadata from the federation endpoint.
   *
   * @param {string} _id - The client_id (URI)
   * @returns {Object | undefined} The client metadata object
   */
  async fetchClientMetadata(_id) {
    let jwtData;
    try {
      const response = await axios.get(`${_id}/.well-known/openid-federation`);
      jwtData = response.data; // Assuming the JWT is returned as plain text
    } catch (error) {
      console.error(`Error fetching JWT from ${_id}:`, error);
      return undefined;
    }

    // Decode the JWT to extract client metadata
    let decodedJwt;
    try {
      decodedJwt = jwt.decode(jwtData);
      if (!decodedJwt) {
        throw new Error('Failed to decode JWT');
      }
    } catch (error) {
      console.error(`Error decoding JWT from ${_id}:`, error);
      return undefined;
    }

    // Extract the metadata specific to 'openid_relying_party'
    const metadata = decodedJwt.metadata?.openid_relying_party;
    if (!metadata) {
      console.error(`Missing openid_relying_party metadata in JWT from ${_id}`);
      return undefined;
    }

    // Construct the client object based on the federation metadata
    return this.constructClientFromMetadata(metadata);
  }

  /**
   * Construct a new client and upsert it into the database.
   *
   * @param {string} _id - The client_id (URI)
   * @returns {Object | undefined} The newly created client object
   */
  async createNewClient(_id) {
    // Fetch new client metadata from the federation
    const newClient = await this.fetchClientMetadata(_id);

    if (!newClient) {
      console.error(`Failed to create new client ${_id} from federation.`);
      return undefined;
    }

    // Add federationCreationTimestamp to the new client metadata
    newClient.federationCreationTimestamp = new Date().toISOString();

    // Save the new client in the database
    try {
      await this.upsert(_id, newClient);
    } catch (error) {
      console.error(`Error saving client ${_id} to DB:`, error);
      return undefined;
    }

    return newClient;
  }

  /**
   * Construct the client object based on the OpenID Federation metadata.
   *
   * @param {Object} metadata - The openid_relying_party metadata
   * @returns {Object} The client object
   */
  constructClientFromMetadata(metadata) {
    const clientId = metadata.client_id;
    const jwksUri = metadata.jwks_uri;

    const contacts = Array.isArray(metadata.contacts) ? metadata.contacts : [];
    const grant_types = Array.isArray(metadata.grant_types) ? metadata.grant_types : [];

    // Construct the client object
    const client = {
      client_id: clientId,
      client_name: metadata.client_name,
      jwks_uri: jwksUri,
      redirect_uris: metadata.redirect_uris,
      grant_types: grant_types,
      contacts: contacts,
      logo_uri: metadata.logo_uri,
      token_endpoint_auth_method: metadata.token_endpoint_auth_method,
      sector_identifier_uri: metadata.sector_identifier_uri,
      software_id: metadata.software_id,
      software_version: metadata.software_version,
      subject_type: metadata.subject_type,
      require_signed_request_object: metadata.require_signed_request_object,
      organization_name: metadata.organization_name,
    };

    // Extract org_id and software_id from the jwks_uri
    const [_, orgId, softwareId] = jwksUri.match(/\/([^/]+)\/([^/]+)\/application\.jwks$/) || [];

    // Add tls_client_auth_subject_dn if org_id and software_id were extracted successfully
    if (orgId && softwareId) {
      client.tls_client_auth_subject_dn = `CN=${softwareId},OU=${orgId},O=Raidiam,C=UK`;
    }

    // Filter out any null values
    return filterNullFields(client);
  }
}

export default ClientsMongoAdapter;
