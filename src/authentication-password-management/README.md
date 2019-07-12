Authentication and Password Management
======================================

[OWASP Secure Coding Practices][1] is a valuable document for programmers to
help them to validate if all best practices were followed during project
implementation. Authentication and Password Management are critical parts of any
system and they are covered in detail from user signup, to credentials storage,
password reset and private resources access.

Some guidelines may be grouped for more in-depth details. Plus, source code 
examples are provided to illustrate the topics.

## Rules of Thumb

Let's start with the rules of thumb: "_all authentication controls must be
enforced on a trusted system_" which usually is the server where the
application's backend is running.

For the sake of system's simplicity, and to reduce the points of failure, you
should utilize standard and tested authentication services. Usually frameworks
already have such a module and you're encouraged to use them as they are
developed, maintained, and used by many people behaving as a centralized
authentication mechanism. Nevertheless, you should "_inspect the code carefully
to ensure it is not affected by any malicious code_", and be sure that it
follows the best practices.

Resources which require authentication should not perform it themselves.
Instead, "_redirection to and from the centralized authentication control_"
should be used. Be careful handling redirection: you should redirect only to
local and/or safe resources.

Authentication should not be used only by the application's users, but also by
your own application when it requires "_connection to external systems that
involve sensitive information or functions_". In these cases, "_authentication
credentials for accessing services external to the application should be
encrypted and stored in a protected location on a trusted system (e.g., the
server). The source code is NOT a secure location_".

[1]: https://www.owasp.org/index.php/OWASP_Secure_Coding_Practices_-_Quick_Reference_Guide
