import datetime
import json
import sys

# Deps in src/vendor
sys.path.insert(0, 'src/vendor')

from firetail_lambda import firetail_handler, firetail_app  # noqa: E402
app = firetail_app()


@firetail_handler(app)
def endpoint(event, context):
    return {
        "statusCode": 200,
        "body": json.dumps({
            "message": "Hello, the current time is %s" % datetime.datetime.now().time()
        })
    }