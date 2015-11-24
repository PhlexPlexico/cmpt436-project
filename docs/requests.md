Requests From Client To Server
=======

### Get total group balance
 * args: groupId
 * return: total group account balance (total of all purchases)

### Get all balances for a group
 * args: groupId
 * return: array of account balances per user in group (amount owed)
 
### Post a purchase (within a group)
 * args: userId, groupId, amount, array of expected split of purchase among group
 * return: successful?
 
### Post a payment (within a group)
 * args: userId, userId, group, amount
 * return: successful?
 
### Get all feed items for a group
 * args: groupId
 * return: list of all the feed items for that group. 
 
### Post a new feed item for a group
 * args: feed item, groupId
 * return: successful?
 
### Put a new user
 * args: email?
 * return: successful?
 
### Put user leaves group
 * args: userId, groupId, array of split of user's balance among group.
 * return: successful?
 
### Put user joins group
 * args: userId, groupId, new user's assigned balance (amount owed)
 * return: successful?
 
### Put create group
 * args: userId, groupId
 * return: successful?