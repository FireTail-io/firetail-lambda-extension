import sys, base64, datetime, json, os, time

# Deps in src/vendor
sys.path.insert(0, 'src/vendor')
import requests

from firetail_lambda import firetail_handler, firetail_app
app = firetail_app()

from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all
patch_all()

@firetail_handler(app)
def endpoint(event, context):
    return {
        "statusCode": 200,
        "body": json.dumps({
            "message": "Hello, the current time is %s" % datetime.datetime.now().time()
        })
    }