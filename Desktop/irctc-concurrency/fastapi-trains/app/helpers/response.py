from fastapi.responses import JSONResponse
def success_response(message: str, data=None):
    return {
        "success": True,"message": message,"data":data}
def error_response(message: str, data=None):
    return {
        "success": False,"message": message,"data":data}