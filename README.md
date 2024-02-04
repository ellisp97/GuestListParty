## Project Task

Implement a guestlist service for a party, the venue is undecided so the number of tables and the capacity are subject to change.

When the party begins, guests will arrive with an entourage. This party may not be the size indicated on the guest list. 
However, if it is expected that the guest's table can accommodate the extra people, then the whole party should be let in. Otherwise, they will be turned away.
Guests will also leave throughout the course of the party. Note that when a guest leaves, their accompanying guests will leave with them.

At any point in the party, we should be able to know:
- Our guests at the party
- How many empty seats there are

### Add a guest to the guestlist

If there is insufficient space at the specified table, then an error should be thrown.

```
POST /guest_list/name
body: 
{
    "table": int,
    "accompanying_guests": int
}
response: 
{
    "name": "string"
}
```

### Get the guest list

```
GET /guest_list
response: 
{
    "guests": [
        {
            "name": "string",
            "table": int,
            "accompanying_guests": int
        }, ...
    ]
}
```

### Guest Arrives

A guest may arrive with an entourage that is not the size indicated at the guest list.
If the table is expected to have space for the extras, allow them to come. Otherwise, this method should throw an error.

```
PUT /guests/name
body:
{
    "accompanying_guests": int
}
response:
{
    "name": "string"
}
```

### Guest Leaves

When a guest leaves, all their accompanying guests leave as well.

```
DELETE /guests/name
```

### Get arrived guests

```
GET /guests
response: 
{
    "guests": [
        {
            "name": "string",
            "accompanying_guests": int,
            "time_arrived": "string"
        }
    ]
}
```

### Count number of empty seats

```
GET /seats_empty
response:
{
    "seats_empty": int
}
```

## Project Solution

#### Setting up the container
I have provided a makefile with the service, to run the docker image issue the `make mysql` command. This will list the container under the name mysql5.7 running with the mysql:5.7 version.

#### Setting up the mysql db
To create an instance of the db which will be named *guestlist_db* issue the `make createdb` command, note this can also be reverted using the `make dropdb` command.
To handle schema migration and to set this up to handle db schema changes in the future the *golang-migrate* library is used. There is only one migrate up and down file and to fill the schema following the createdb command issue a `make migrateup`, note likewise this can also be reverted using a `make migratedown` command. At this point the container running should allow the connections to the database with the _guests_,_arrivals_ and _tables_ schemas.

#### Start
Finally to start the sever do `make server`

## Testing

I've provided multiple types of unit testing, firstly there database CRUD functions to test the mysql queries I have set up. This also uses the *sqlc* package which is used to generate the .sql.go files from the user defined queries (db/query/). Secondly the api functions exposed using gin are fully mocked using the *gomock* package this will allow for faster, cleaner tests which dont have to rely on the db connections this has a 99% coverage for all functions exposed to the user. 
Both of these can be run by using the `make test` command, while the mocked tests will produce fake test data, the mysql tests will produce real data viewable in the database.

## Documentation

Additional documentation is provided using the *swagger* package this produces a auto-generated frontend based on function comments which is viewable while the server is running at `http://localhost:3000/swagger/index.html#/`.

## Utility
All the configuration are parsed from the app.env file, if any config changes are neccessary.
Given more time there may be scope to parallelise the tests and introduce some waitGroup concepts.