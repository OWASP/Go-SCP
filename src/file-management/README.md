File Management
===============

The first precaution to take when handling files is to make sure the users are
not allowed to directly supply data to any dynamic functions. In languages like
PHP, passing user data to _dynamic include_ functions, is a serious security
risk. Go is a compiled language, which means there are no `include` functions,
and libraries aren't usually loaded dynamically[^1].

File uploads should only be permitted from authenticated users.
After guaranteeing that file uploads are only made by authenticated users,
another important aspect of security is to make sure that only acceptable
file types can be uploaded to the server (_whitelisting_).
This check can be made using the following Go function that detects MIME types:
`func DetectContentType(data []byte) string`

Below you find the relevant parts of a simple program to read and compute
filetype ([filetype.go][0])

```go
{...}
// Write our file to a buffer
// Why 512 bytes? See http://golang.org/pkg/net/http/#DetectContentType
buff := make([]byte, 512)

_, err = file.Read(buff)
{...}
//Result - Our detected filetype
filetype := http.DetectContentType(buff)
```

After identifying the filetype, an additional step is required to validate the
filetype against a whitelist of allowed filetypes. In the example, this is
achieved in the following section:

```go
{...}
switch filetype {
case "image/jpeg", "image/jpg":
  fmt.Println(filetype)
case "image/gif":
  fmt.Println(filetype)
case "image/png":
  fmt.Println(filetype)
default:
  fmt.Println("unknown file type uploaded")
}
{...}
```

Files uploaded by users should not be stored in the web context of the
application. Instead, files should be stored in a content server or in a
database. An important note is for the selected file upload destination not to
have execution privileges.

If the file server that hosts user uploads is \*NIX based, make sure to
implement safety mechanisms like chrooted environment, or mounting the target
file directory as a logical drive.

Again, since Go is a compiled language, the usual risk of uploading files that
contain malicious code that can be interpreted on the server-side, is
non-existent.

In the case of dynamic redirects, user data should not be passed. If it is
required by your application, additional steps must be taken to keep the
application safe. These checks include accepting only properly validated data
and relative path URLs.

Additionally, when passing data into dynamic redirects, it is important to make
sure that directory and file paths are mapped to indexes of pre-defined lists
of paths, and to use these indexes.

Never send the absolute file path to the user, always use relative paths.

Set the server permissions regarding the application files and resources to
`read-only`. And when a file is uploaded, scan the file for viruses and malware.

[^1]:  Go 1.8 does allow dynamic loading now, via [the new plugin mechanism]( https://golang.org/pkg/plugin/).
       If your application uses this mechanism, you should take precautions
       against user-supplied input.

[0]: ./filetype/filetype.go
