import base64, datetime, json, os, time

def firetail_logger(token):
    def decorator(func):
        def wrapper_func(*args, **kwargs):
            start_time = time.time()

            # Unpack the args
            event, _ = args

            # Get the response returned down the chain
            response = func(*args, **kwargs)

            # Create our log payload, and print it
            log_payload = base64.b64encode(json.dumps({"event": event,"response": response}).encode("utf-8")).decode("ascii")
            print("firetail:%s:%s" % (token, log_payload))

            # Ensure the execution time is >25ms to give the logs API time to propagate our print() to the extension.
            time.sleep(max(time.time() - start_time + 25/1000, 0))

            # Return the response from down the chain
            return response
        return wrapper_func
    return decorator

@firetail_logger(os.getenv("FIRETAIL_TOKEN"))
def endpoint(event, context):
    return {
        "statusCode": 200,
        "body": json.dumps({
            "message": "Hello, the current time is %s" % datetime.datetime.now().time()
        })
    }