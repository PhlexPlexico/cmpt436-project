Basic Structure
=======

UI
-----------
 
### Homepage

The bulk of the screen is probably some kind of Feed tailored to this user.

Left sidebar:
 * contacts: list of contacts added by this user. A zero-sum balance is maintained between the user and each contact. Contacts are real users who have their own account.
 * actors (name?): list of *non-user* entities added by this user. A zero-sum balance is maintained between the user and each actor. (A user can make up any actor they want, and has complete control over the balance.)
 * groups: list of payment groups this user is a part of. 

Clicking on the link to any particular contact, actor, or Group takes the user to that particular Feed Page.


### Feed Page.

Feed Pages are identical for contacts, actors and groups. They just consist of a Feed of past interactions with this contact, actor, or Group, as well as a list of balances for each actor/contact involved.


### Feed

A Feed is a chronological list of Posts, with an option to enter new messages at the bottom.


### Post

A Post is a message on a Feed. There are two kinds of posts: transactions, and messages. 
  * Transactions are added to the Feed whenever the account balance of the Group is affected. Users can comment on transactions. 
  * Messages are either basic user chat messages or system notifications (e.g. "Josh has entered the Group"). Users cannot comment on these.

### Transactions

There are two kinds of transactions: purchases and repayments. 

A purchase occurs when a user changes the balance for a whole Group (for instance by paying for a hotel). Negative purchases are also allowed (e.g. someone gave the Group 100$), which is why I'd like to think of a word other than 'purchase'. When a user enters a purchase, they must decide how to split the new money/costs among the Group (a default even split is provided).

A repayment occurs when one user indicates they have repaid another user. This simply changes their respective account balances within the Group.



Modelling Interactions
---------------------

### 'Groups' simplification

All interactions are represented by Groups internally. All interactions, whether between two contacts, or between a user and actors, or between multiple users in an 'actual' Group, are modeled by Groups. This decreases the headache on the back-end, and makes the distinction between these things purely a UI problem.


### 'Actors' simplification

Outside of Groups (i.e. outside of any financial interactions), the only user entities are real users, with real accounts. Within groups, the only entities are Actors. Real users are just modeled as Actors, and 'actual' Actors (fake users) are also Actors. Besides simplifying Group interactions, this makes it easy to include within a Group a person who does not yet have an account. When the real user joins the Group, they can then claim the Actor that has been assigned to them by the other users.


Modelling Transactions
---------------------

### Group balance simplification

Whenever an actor leaves the Group, a decision is made about how to split the outgoing actor's balance.

Whenever an actor enters the Group, a decision is made about how much of the current balance to allocate to the actor.

Whenever a new purchase occurs, a decision is made about how to split the expected contributions among the actors.


### Transactions Database representation

Within the database, each Group stores a total balance. 

It also stores two quantities for each actor within the Group: that actor's expected contribution, and their actual contribution, to the Group's total balance. With these two numbers, it is easy to calculate an actor's current balance:

  balance = expected contribution - actual contribution

This representation also makes it easy to update everybody's individual balances when somebody leaves or joins the Group.


Other Notes
------------

At least one admin user is assigned for each Group. An admin can:
  * Approve a user/actor's departure or entry into the current Group. This entails deciding how to change the Group account balance after the departure/entry. 
  * Approve new transactions. For purchases, this entails deciding how to split the new cost among the Group members. 
  * More to come.
