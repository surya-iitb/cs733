# cs733

The server has following functionalities:

1. Writes a file to server with or without without expiry time. The file contents are updated if it laready exists.

2. Takes read request and sent the contents or the error message to the client.

3. Takes cas(compare and swap) request. Updates the version and the contents, if the version number matches. Updated vesion number is sent to the client.

4. Delete a file, if it alreay exists. Send an error otherwise.

----------------------------------------------------------------------------------

Concurrency Control- The server implements lock in read and write modes. The operations have to wait if a write or read is already in progress.

----------------------------------------------------------------------------------

The expiry time is also handled.


