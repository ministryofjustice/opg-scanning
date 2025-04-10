flowchart TD
%% API Endpoints
subgraph API [API Endpoints]
A1[HTTP Request: /health-check]
A2[HTTP Request: /auth/sessions]
A3[HTTP Request: /api/ddc]
end

%% Index Controller Handlers
subgraph IndexController [IndexController]
H1[Health Check Handler]
H2[Auth Handler]
H3[IngestHandler]
H4[HTTP Rresponse: xxx]
end

A1 --> H1
A2 --> H2
A3 --> H3

%% Ingest Flow
subgraph IngestFlow [Ingestion & Processing]
I1[Read Request Body]
I2[Validate Content-Type]
I3[Extract Schema & Validate XSD]
I4[XML & Embedded Document Validation]
I5[Validate Set via Validator]
I6[Create Case Stub Sirius]
I7[Process Each Document]
end

H3 --> I1
I1 --> I2
I2 --> I3
I3 --> I4
I4 --> I5
I5 --> I6
I6 --> I7
IngestFlow --> H4

%% Job Queue & Document Processing
subgraph JobProcessing [Sequential JobProcessing]
Q1[DocumentProcessor]
Q2[Component Registry]
Q3[Parser / Validator per Doc Type]
Q4[Process Document]
Q5[On-Complete Callback]
end

I7 --> Q1
Q1 --> Q2
Q2 --> Q3
Q3 --> Q4
Q4 --> Q5
Q5 --> onComplete

subgraph onComplete [onComplete callback]
Q8[Attach Documents to Case]
Q9[Persist Data to AWS S3]
Q10[Queue for External AWS Processing]
end

Q8 --> Q9
Q9 --> Q10
onComplete --> JobProcessing
JobProcessing --> IngestFlow

%% Auth Flow
subgraph AuthFlow [Authentication]
AU1[Authenticate Credentials]
AU2[Generate JWT Token]
AU3[Return Auth Response]
end

H2 --> AU1
AU1 --> AU2
AU2 --> AU3

%% External Systems
subgraph External [External Systems]
S1[Sirius Case Management]
S2[AWS S3]
S3[AWS Queue]
end

I6 -- Create Case Stub --> S1
Q8 -- Attach to Case Stub --> S1

Q9 -- Persist Form Data --> S2
Q10 -- Send for Processing --> S3

%% Optional: Health check can be a simple endpoint returning OK
H1 -- Return OK --> A1

%% Styling (optional)
classDef main fill:#e0f7fa,stroke:#00796b,stroke-width:2px;
class A1,A2,A3,H1,H2,H3,AU1,AU2,AU3 main;
