# Copyright (C) 2020 - 2021 Divkix. All rights reserved. Source code available under the AGPL.
#
# This file is part of Alita_Robot.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

from pymongo import MongoClient

from alita import DB_NAME, DB_URI, LOGGER

alita_db_client = MongoClient(DB_URI)
alita_main_db = alita_db_client[DB_NAME]


class MongoDB:
    """Class for interacting with Bot database."""

    def __init__(self, collection) -> None:
        self.collection = alita_main_db[collection]

    # Insert one entry into collection
    def insert_one(self, document):
        result = self.collection.insert_one(document)
        return repr(result.inserted_id)

    # Find one entry from collection
    def find_one(self, query):
        result = self.collection.find_one(query)
        if result:
            return result
        return False

    # Find entries from collection
    def find_all(self, query=None):
        if query is None:
            query = {}
        return list(self.collection.find(query))

    # Count entries from collection
    def count(self, query=None):
        if query is None:
            query = {}
        return self.collection.count_documents(query)

    # Delete entry/entries from collection
    def delete_one(self, query):
        self.collection.delete_many(query)
        return self.collection.count_documents({})

    # Replace one entry in collection
    def replace(self, query, new_data):
        old = self.collection.find_one(query)
        _id = old["_id"]
        self.collection.replace_one({"_id": _id}, new_data)
        new = self.collection.find_one({"_id": _id})
        return old, new

    # Update one entry from collection
    def update(self, query, update):
        result = self.collection.update_one(query, {"$set": update})
        new_document = self.collection.find_one(query)
        return result.modified_count, new_document

    @staticmethod
    def close():
        return alita_db_client.close()


def __connect_first():
    _ = MongoDB("test")
    LOGGER.info("Initialized Database!\n")


__connect_first()
