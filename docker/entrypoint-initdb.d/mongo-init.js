// Switch to the 'alita' database
db = db.getSiblingDB('alita');

// create a new user
db.createUser({
    user: "admin",
    pwd: "admin",
    roles: [
        {
            role: "readWrite",
            db: "alita"
        }
    ]
});
