# Initially no messages are returned
# GET all messages, should be empty initially
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

# Trying to like the same message again should return a conflict (409)
POST http://localhost:8080/messages/{{message_id}}/reactions
{ "type": "like", "user_id": "testuser" }
HTTP 409

# Trying to clap the same message again should return a conflict (409)
POST http://localhost:8080/messages/{{message_id}}/reactions
{ "type": "clap", "user_id": "testuser" }
HTTP 409

#Invalid reaction type gives validation error
POST http://localhost:8080/messages/{{message_id}}/reactions
{ "type": "dislike", "user_id": "testuser" }
HTTP 400

# Add a new message
POST http://localhost:8080/messages
{ "text": "another message!", "user_id": "anotheruser" }
HTTP 201

# Capture new message ID
GET http://localhost:8080/messages
HTTP 200
[Captures]
new_message_id: jsonpath "$.messages[0].id"

# Add "like" reactions from 4 different users
POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "like", "user_id": "user1" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "like", "user_id": "user2" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "like", "user_id": "user3" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "like", "user_id": "user4" }
HTTP 201

# Add "clap" reactions from 4 different users
POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "clap", "user_id": "cuser1" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "clap", "user_id": "cuser2" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "clap", "user_id": "cuser3" }
HTTP 201

POST http://localhost:8080/messages/{{new_message_id}}/reactions
{ "type": "clap", "user_id": "cuser4" }
HTTP 201

# Verify the latest message has 4 likes and 4 claps
GET http://localhost:8080/messages
HTTP 200
[Asserts]
jsonpath "$.messages[?(@.id == '{{new_message_id}}')].message_reactions[?(@.type == 'like')].count" nth 0 == 4
jsonpath "$.messages[?(@.id == '{{new_message_id}}')].message_reactions[?(@.type == 'clap')].count" nth 0 == 4

# Add 10 more messages
POST http://localhost:8080/messages
{ "text": "message 1", "user_id": "user1" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 2", "user_id": "user2" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 3", "user_id": "user3" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 4", "user_id": "user4" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 5", "user_id": "user5" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 6", "user_id": "user6" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 7", "user_id": "user7" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 8", "user_id": "user8" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 9", "user_id": "user9" }
HTTP 201

POST http://localhost:8080/messages
{ "text": "message 10", "user_id": "user10" }
HTTP 201

# Capture the ID of the last message
GET http://localhost:8080/messages
HTTP 200
[Captures]
last_message_id: jsonpath "$.messages[0].id"

# Add 3 likes to the last message
POST http://localhost:8080/messages/{{last_message_id}}/reactions
{ "type": "like", "user_id": "likeuser1" }
HTTP 201

POST http://localhost:8080/messages/{{last_message_id}}/reactions
{ "type": "like", "user_id": "likeuser2" }
HTTP 201

POST http://localhost:8080/messages/{{last_message_id}}/reactions
{ "type": "like", "user_id": "likeuser3" }
HTTP 201

# Verify the last message has 3 likes
GET http://localhost:8080/messages
HTTP 200
[Asserts]
jsonpath "$.messages[?(@.id == '{{last_message_id}}')].message_reactions[?(@.type == 'like')].count" nth 0 == 3

# Fetch the second page of messages
GET http://localhost:8080/messages?page=2
HTTP 200
[Asserts]
jsonpath "$.messages" count == 3