# Note: This assumes empty databases
# You can wipe the data in Docker with docker compose down && docker compose up -d

# Initially no messages are returned

GET http://localhost:8080/messages
HTTP 200
[Asserts]
jsonpath "$.messages" count == 0

# Insert a few messages

POST http://localhost:8080/messages
{ "text": "hello", "user_id": "testuser" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "world!", "user_id": "testuser" }
HTTP 201

# Now, we can get back 2 messages

GET http://localhost:8080/messages
HTTP 200
[Captures]
message_id: jsonpath "$.messages[0].id"
[Asserts]
jsonpath "$.messages" count == 2

# The messages are sorted by the time they were created in descending order
jsonpath "$.messages[0].text" == "world!"

# Create a reaction to the latest message with default value #TODO: ReplaceIDs
POST http://localhost:8080/messages/a5922418-ab74-4a9f-ade0-c03754933265/reactions
{ "type": "clap", "user_id": "testusse7" }
HTTP 201

# React to the same message again and its a conflict
POST http://localhost:8080/messages/a5922418-ab74-4a9f-ade0-c03754933265/reactions
{ "type": "like", "user_id": "xa5ssr9" }
HTTP 409

# React to the message as testuser2 with score
POST http://localhost:8080/messages/f349e150-91d7-40b7-86c6-41e7615196a0/reactions
{ "type": "clap", "user_id": "testssuser2", "score": 5 }
HTTP 201

GET http://localhost:8080/messages?page=2