from pymongo import MongoClient

# Setup a test MongoDB client and collection
client = MongoClient("mongodb://localhost:27017")
db = client["signals_db"]
collection = db["streams"]

# Clean up any pre-existing data for a clean test run
collection.delete_many({"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"})

# Insert the initial stream configuration as per the test setup
initial_data = {
    "stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f",
    "events_supported": ["event1", "event2"],
    "events_endpoint": "http://example.com/events",
    "status": "enabled",
    "subjects": []
}
insert_result = collection.insert_one(initial_data)
print("Insertion Result:", insert_result)

# Retrieve the inserted document to verify it matches what was expected
retrieved_data = collection.find_one({"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"})
print("Retrieved Data:", retrieved_data)

# Attempt to update the status to 'paused' as the test would
update_result = collection.find_one_and_update(
    {"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"},
    {"$set": {"status": "paused", "reason": "Maintenance"}},
    return_document=True
)
print("Update Result:", update_result)

# Fetch the updated record to verify changes
updated_data = collection.find_one({"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"})
print("Updated Data:", updated_data)

# Clean up after the test
collection.delete_many({"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"})
client.close()