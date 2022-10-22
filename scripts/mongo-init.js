print('START MONGO INIT #################################################################');
db = db.getSiblingDB('gately_store');
db.createUser(
    {
        user: "user",
        pwd: "pass",
        roles: [
            {
                "role": "readWrite",
                "db": "gately_store"
            }
        ]
    }
);

db.createCollection( "url_mappings" ,
    {"$jsonSchema":{"bsonType":"object","additionalProperties":false,"required":["_id"],
            "properties": {
                "_id":{"bsonType":"string","description":"PK shortUrl becomes _id "},
                "longUrl":{"bsonType":"string"},
                "hits":{"bsonType":"int"},
                "createdTs":{"bsonType":"int"}}},
        "$expr": [
        ]}
);

print('END MONGO INIT #################################################################');
