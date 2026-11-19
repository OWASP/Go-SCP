Output Encoding
===============

Although output encoding only has six bullets in the section on [OWASP SCP Quick
Reference Guide][1], undesirable practices of Output Encoding are rather
prevalent in Web Application development, thus leading to the Top 1
vulnerability: [Injection][2].

As Web Applications become more complex, the more data sources they usually
have, for example: users, databases, thirty party services, etc. At some point
in time collected data is outputted to some media (e.g. a web browser) which has
a specific context. This is exactly when injections happen if you do not have a
strong Output Encoding policy.

Certainly you've already heard about all the security issues we will approach
in this section, but do you really know how do they happen and/or how to avoid
them?

[1]: https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/
[2]: https://owasp.org/Top10/A03_2021-Injection/
