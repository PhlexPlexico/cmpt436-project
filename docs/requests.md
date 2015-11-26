Requests From Client To Server
=======

Preliminary
---------


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

### Websocket: get constantly updating feed items for group (including bulk get for historical items)
 * args: groupId
 * return: constantly updating list of all the feed items for that group. 
   * This could also be used to infer constantly updating balances, and constantly updating system
     notifications, such as users entering/leaving group. Come to think of it, the websocket could 
     cover a lot of stuff.

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
 * args: userId, groupId, new user's assigned balance (amount owed), array of changes to other users' balances
 * return: successful?

### Put create group
 * args: userId, group name
 * return: successful?
