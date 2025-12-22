# Common scanning problems

These are some common things that might go wrong in the scanning ecosystem, and how they can be resolved.

## Document stuck in processing

### Symptoms

A document may get locked in a processing state. This presents itself in the logs with this error message:

> failed to set document to processing '{{ Document ID }}': ConditionalCheckFailedException: The conditional request failed

When a document is ingested, a DynamoDB lock record is created with the Document ID and a status of "PROCESSING". Upon completion, the status is either updated to "COMPLETED" or "FAILED". Any subsequent request with the same ID will be rejected, unless the status is FAILED (it is valid behaviour to retry a failed scan upload).

The document can get stuck in a PROCESSING state if it never completes: this is most likely because the request has timed out or panicked and been cancelled before the final status is set.

### Resolution

You need to use the Document ID to query logs and identify the document being submitted. You then need to cross-reference with Sirius to see if the upload was successful or not. You then need to update the DynamoDB record (using a role that can write data) to reflect the actual status of the upload.

## Document size too big

### Symptoms

HTTP requests from our scanning supplier may be too large and get rejected with a 431 error. This appears in the logs with the message:

> Request content too large: the XML document exceeds the maximum allowed size

This typically occurs when scanning a particularly long document (e.g. 100s of pages) that has been sent as a single large PDF file.

### Resolution

This is automatically reported back through the supplier's technology and users will be prompted to separate the upload into multiple smaller payloads.

## Sirius Dead Letter Queue

### Symptoms

When form data has been extracted from a scanning upload and fails to be processed by Sirius, it will be added to the Dead Letter Queue (DLQ). As soon as anything is added to the queue, it triggers an alarm for whoever is on-call for Sirius.

### Resolution

After retrieving the document ID from the DLQ message, search the Sirius logs to identify why the upload failed. The resolution then varies depending on the exact circumstances.

#### Temporary issues

If the cause of the failure was temporary, for example because a network resource was unavailable or an unrelated issue with Sirius that has been resolved, the messages can be retried by using the AWS console to redrive them onto the normal DDC queue.

#### Data too long for field

When an individual piece of data is too big for the database field it's trying to be inserted into, Sirius will fail with the log message:

> An exception occurred while executing a query: SQLSTATE[22001]: String data, right truncated: 7 ERROR: value too long for type character varying(255)

Because the data sent from the scanning supplier is invalid, this issue cannot be automatically resolved.

You must extract the Document ID via logging and report it to the OPG Scanning team so that they can rescan the document. It can also be helpful to retrieve the related XML file from the jobsqueue bucket (using a role that can read data) to identify and report back the specific fields which were too large.
