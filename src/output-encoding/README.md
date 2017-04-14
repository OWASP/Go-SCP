Output Encoding
===============

Although only has a six bullets only section on
[OWASP SCP Quick Reference Guide][1], bad practices of Output Encoding are
pretty prevalent on Web Application development, thus leading to the Top 1
vulnerability: [Injection][2].

As complex and rich as Web Applications become, the more data sources they have:
users, databases, thirty party services, etc. At some point in time collected
data is outputted to some media (eg. web browser) which has a specific context.
This is exactly when injections happen if you do not have a strong Output
Encoding policy.

Certainly you have already heard about all the security issues we will approach
in this section, but do you really know how do they happen and/or how to avoid
them?

[1]: https://www.owasp.org/images/0/08/OWASP_SCP_Quick_Reference_Guide_v2.pdf
[2]: https://www.owasp.org/index.php/Top_10_2013-A1-Injection
