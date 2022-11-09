import base64, datetime, json, os, time

def firetail_logger(logging_uuid):
    def decorator(func):
        def wrapper_func(*args, **kwargs):
            start_time = time.time()

            # Unpack the args
            event, _ = args

            # Get the response returned down the chain
            response = func(*args, **kwargs)

            # Create our log payload, and print it
            log_payload = base64.b64encode(json.dumps({"event": event,"response": response}).encode("utf-8")).decode("ascii")
            print("firetail:%s:%s" % (logging_uuid, log_payload))

            # Ensure the execution time is >25ms to give the logs API time to propagate our print() to the extension.
            time.sleep(max(500/1000 - (time.time() - start_time), 0))

            # Return the response from down the chain
            return response
        return wrapper_func
    return decorator

@firetail_logger(os.getenv("FIRETAIL_LOGGING_UUID"))
def endpoint(event, context):
    return {
        "statusCode": 200,
        "body": json.dumps({
            "message": "Hello, the current time is %s" % datetime.datetime.now().time()
        })
    }