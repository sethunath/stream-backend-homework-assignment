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

# Create a reaction to the latest message
POST http://localhost:8080/messages/{{message_id}}/reactions
{ "type": "like", "user_id": "testuser" }
HTTP 201

