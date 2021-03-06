# cs733

The server has following functionalities:

1. Takes read or write requests with or without without expiry time. The file contents are updated if it laready exists.

2. Takes read request and sent the contents or the error message to the client.

3. Takes cas(compare and swap) request. Updates the version and the contents, if the version number matches. Updated vesion number is sent to the client.

4. Delete a file, if it alreay exists. Send an error otherwise.

----------------------------------------------------------------------------------

Concurrency Control- The server implements lock in read and write modes. The operations have to wait if a write or read is already in progress. Concurrency is not handled at the file level. All the clients and the threads have to wait, if a single file is being written by other threads of some other client.

----------------------------------------------------------------------------------

Test File
------------

In the test file, some basic test cases are covered.

1. Writing to a file

2. Reading from the written file after 5 seconds. The expiry time left is checked.

3. Reading a non-existent file.

4. Comparing and swapping an existing file with correct version number.

5. Comparing and swapping a non-existent file.

6. Comparing and swapping with version number mismatch.

7. Delete an existing file.

8. Read from a non-existent file.
