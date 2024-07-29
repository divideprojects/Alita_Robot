// Switch to the 'alita' database
// this is the database that we want to create the user in and store files in
db = db.getSiblingDB('alita');

// create a new basic user with readWrite role on the 'alita' database
// the user is 'admin' with password 'admin'
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
