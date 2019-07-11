Other guidelines
================

Authentication is a critical part of any system, therefore you should always
employ correct and safe practices. Below are some guidelines to make your
authentication system more resilient:

* "_Re-authenticate users prior to performing critical operations_"
* "_Use Multi-Factor Authentication for highly sensitive or high value
  transactional accounts_"
* "_Implement monitoring to identify attacks against multiple user accounts,
  utilizing the same password. This attack pattern is used to bypass standard
  lockouts, when user IDs can be harvested or guessed_"
* "_Change all vendor-supplied default passwords and user IDs or disable the
  associated accounts_"
* "_Enforce account disabling after an established number of invalid login
  attempts (e.g., five attempts is common).  The account must be disabled for a
  period of time sufficient to discourage brute force guessing of credentials,
  but not so long as to allow for a denial-of-service attack to be performed_"
