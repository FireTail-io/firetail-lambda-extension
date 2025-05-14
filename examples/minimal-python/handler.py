import datetime
import json
import sys

# Deps in src/vendor
sys.path.insert(0, "src/vendor")


def endpoint(event, context):
    current_time = datetime.datetime.now().time()
    return {
        "statusCode": 200,
        "body": json.dumps({"message": "Hello, the current time is %s" % current_time}),
        "headers": {"Current-Time": "%s" % current_time},
    }
