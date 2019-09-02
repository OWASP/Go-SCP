Communication Security
======================

When approaching communication security, developers should be certain that the
channels used for communication are secure.
Types of communication include server-client, server-database, as well as all
backend communications. These must be encrypted to guarantee data integrity, and
to protect against common attacks related to communication security.
Failure to secure these channels allows known attacks like MITM, which allows
attacker to intercept and read the traffic in these channels.

The scope of this section covers the following communication channels:

* HTTP/HTTPS
* Websockets

[1]: https://www.owasp.org/index.php/Man-in-the-middle_attack
