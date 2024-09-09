// ClientsMongoAdapter.js

import axios from 'axios';
import jwt from 'jsonwebtoken';
import MongoAdapter from './mongodb.js'; // Adjust the path as necessary

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
   * Override the find method to implement find-or-fetch-create logic.
   *
   * @param {string} _id - The client_id (URI)
   * @returns {Object | undefined} The client payload
   */
  async find(_id) {
    // Step 1: Look up the client in MongoDB
    const existingClient = await super.find(_id);

    if (existingClient) {
      return existingClient;
    }

    // Step 2: Fetch the JWT from the client_id URI
    let jwtData;
    try {
        const response = await axios.get(`${_id}/.well-known/openid-federation`);
        jwtData = response.data; // Assuming the JWT is returned as plain text
    } catch (error) {
      console.error(`Error fetching JWT from ${_id}:`, error);
      return undefined;
    }

    // Step 3: Decode the JWT to extract client metadata
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

    // Optional: Verify the JWT signature and claims here
    // TODO: Add JWT verification using the issuer's public keys if necessary

    // Extract the metadata specific to 'openid_relying_party'
    const metadata = decodedJwt.metadata?.openid_relying_party;
    if (!metadata) {
      console.error(`Missing openid_relying_party metadata in JWT from ${_id}`);
      return undefined;
    }

    // Step 4: Construct the client object based on the federation metadata
    const newClient = this.constructClientFromMetadata(metadata);

    // Step 5: Save the client in the database for future use
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
    return filterNullFields(client);  }
}

export default ClientsMongoAdapter;