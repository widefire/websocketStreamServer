package RTSP

//method
const (
	//                                           direction       	object  requirement
	RTSP_Method_DESCRIBE      = "DESCRIBE"      //C->S 				P,S		recommended
	RTSP_Method_ANNOUNCE      = "ANNOUNCE"      //C->S, S->C		P,S		optional
	RTSP_Method_GET_PARAMETER = "GET_PARAMETER" //C->S, S->C		P,S		optional
	RTSP_Method_OPTIONS       = "OPTIONS"       //C->S, S->C		P,S		required (S->C: optional)
	RTSP_Method_PAUSE         = "PAUSE"         //C->S				P,S		recommended
	RTSP_Method_PLAY          = "PLAY"          //C->S				P,S		required
	RTSP_Method_RECORD        = "RECORD"        //C->S				P,S		optional
	RTSP_Method_REDIRECT      = "REDIRECT"      //S->C				P,S		optional
	RTSP_Method_SETUP         = "SETUP"         //C->S				S		required
	RTSP_Method_SET_PARAMETER = "PARAMETER"     //C->S, S->C		P,S		optional
	RTSP_Method_SET_TEARDOWN  = "TEARDOWN"      //C->S				P,S		required

)

//status code

const (
	RTSP_Status_Continue                            int = 100
	RTSP_Status_OK                                  int = 200
	RTSP_Status_Created                             int = 201
	RTSP_Status_Low_on_Storage_Space                int = 250
	RTSP_Status_Multiple_Choices                    int = 300
	RTSP_Status_Moved_Permanently                   int = 301
	RTSP_Status_Moved_Temporarily                   int = 302
	RTSP_Status_See_Other                           int = 303
	RTSP_Status_Not_Modified                        int = 304
	RTSP_Status_Use_Proxy                           int = 305
	RTSP_Status_Bad_Request                         int = 400
	RTSP_Status_Unauthorized                        int = 401
	RTSP_Status_Payment_Required                    int = 402
	RTSP_Status_Forbidden                           int = 403
	RTSP_Status_Not_Found                           int = 404
	RTSP_Status_Method_Not_Allowed                  int = 405
	RTSP_Status_Not_Acceptable                      int = 406
	RTSP_Status_Proxy_Authentication_Required       int = 407
	RTSP_Status_Request_Time_out                    int = 408
	RTSP_Status_Gone                                int = 410
	RTSP_Status_Length_Required                     int = 411
	RTSP_Status_Precondition_Faile                  int = 412
	RTSP_Status_Request_Entity_Too_Large            int = 413
	RTSP_Status_Request_URI_Too_Large               int = 414
	RTSP_Status_Unsupported_Media_Type              int = 415
	RTSP_Status_Parameter_Not_Understood            int = 451
	RTSP_Status_Conference_Not_Found                int = 452
	RTSP_Status_Not_Enough_Bandwidth                int = 453
	RTSP_Status_Session_Not_Found                   int = 454
	RTSP_Status_Method_Not_Valid_in_This_State      int = 455
	RTSP_Status_Header_Field_Not_Valid_for_Resource int = 456
	RTSP_Status_Invalid_Range                       int = 457
	RTSP_Status_Parameter_Is_Read_Only              int = 458
	RTSP_Status_Aggregate_operation_not_allowed     int = 459
	RTSP_Status_Only_aggregate_operation_allowed    int = 460
	RTSP_Status_Unsupported_transport               int = 461
	RTSP_Status_Destination_unreachable             int = 462
	RTSP_Status_Internal_Server_Error               int = 500
	RTSP_Status_Not_Implemented                     int = 501
	RTSP_Status_Bad_Gateway                         int = 502
	RTSP_Status_Service_Unavailable                 int = 503
	RTSP_Status_Gateway_Time_out                    int = 504
	RTSP_Status_RTSP_Version_not_supported          int = 505
	RTSP_Status_Option_not_supported                int = 551
)

var statusCodeDescMap map[int]string

func init() {
	statusCodeDescMap = make(map[int]string)
	statusCodeDescMap[RTSP_Status_Continue] = "Continue"
	statusCodeDescMap[RTSP_Status_OK] = "OK"
	statusCodeDescMap[RTSP_Status_Created] = "Created"
	statusCodeDescMap[RTSP_Status_Low_on_Storage_Space] = "Low on Storage Space"
	statusCodeDescMap[RTSP_Status_Multiple_Choices] = "Multiple Choices"
	statusCodeDescMap[RTSP_Status_Moved_Permanently] = "Moved Permanently"
	statusCodeDescMap[RTSP_Status_Moved_Temporarily] = "Moved Temporarily"
	statusCodeDescMap[RTSP_Status_See_Other] = "See Other"
	statusCodeDescMap[RTSP_Status_Not_Modified] = "Not Modified"
	statusCodeDescMap[RTSP_Status_Use_Proxy] = "Use Proxy"
	statusCodeDescMap[RTSP_Status_Bad_Request] = "Bad Request"
	statusCodeDescMap[RTSP_Status_Unauthorized] = "Unauthorized"
	statusCodeDescMap[RTSP_Status_Payment_Required] = "Payment Required"
	statusCodeDescMap[RTSP_Status_Forbidden] = "Forbidden"
	statusCodeDescMap[RTSP_Status_Not_Found] = "Not Found"
	statusCodeDescMap[RTSP_Status_Method_Not_Allowed] = "Method Not Allowed"
	statusCodeDescMap[RTSP_Status_Not_Acceptable] = "Not Acceptable"
	statusCodeDescMap[RTSP_Status_Proxy_Authentication_Required] = "Proxy Authentication Required"
	statusCodeDescMap[RTSP_Status_Request_Time_out] = "Request Timeout"
	statusCodeDescMap[RTSP_Status_Gone] = "Gone"
	statusCodeDescMap[RTSP_Status_Length_Required] = "Length Required"
	statusCodeDescMap[RTSP_Status_Precondition_Faile] = "Precondition Failed"
	statusCodeDescMap[RTSP_Status_Request_Entity_Too_Large] = "Request Entity Too Large"
	statusCodeDescMap[RTSP_Status_Request_URI_Too_Large] = "Request-URI Too Long"
	statusCodeDescMap[RTSP_Status_Unsupported_Media_Type] = "Unsupported Media Type"
	statusCodeDescMap[RTSP_Status_Parameter_Not_Understood] = "Invalid parameter"
	statusCodeDescMap[RTSP_Status_Conference_Not_Found] = "Illegal Conference Identifier"
	statusCodeDescMap[RTSP_Status_Not_Enough_Bandwidth] = "Not Enough Bandwidth"
	statusCodeDescMap[RTSP_Status_Session_Not_Found] = "Session Not Found"
	statusCodeDescMap[RTSP_Status_Method_Not_Valid_in_This_State] = "Method Not Valid In This State"
	statusCodeDescMap[RTSP_Status_Header_Field_Not_Valid_for_Resource] = "Header Field Not Valid"
	statusCodeDescMap[RTSP_Status_Invalid_Range] = "Invalid Range"
	statusCodeDescMap[RTSP_Status_Parameter_Is_Read_Only] = "Parameter Is Read-Only"
	statusCodeDescMap[RTSP_Status_Aggregate_operation_not_allowed] = "Aggregate Operation Not Allowed"
	statusCodeDescMap[RTSP_Status_Only_aggregate_operation_allowed] = "Only Aggregate Operation Allowed"
	statusCodeDescMap[RTSP_Status_Unsupported_transport] = "Unsupported Transport"
	statusCodeDescMap[RTSP_Status_Destination_unreachable] = "Destination Unreachable"
	statusCodeDescMap[RTSP_Status_Internal_Server_Error] = "Internal Server Error"
	statusCodeDescMap[RTSP_Status_Not_Implemented] = "Not Implemented"
	statusCodeDescMap[RTSP_Status_Bad_Gateway] = "Bad Gateway"
	statusCodeDescMap[RTSP_Status_Service_Unavailable] = "Service Unavailable"
	statusCodeDescMap[RTSP_Status_Gateway_Time_out] = "Gateway Timeout"
	statusCodeDescMap[RTSP_Status_RTSP_Version_not_supported] = "RTSP Version Not Supported"
	statusCodeDescMap[RTSP_Status_Option_not_supported] = "Option not support"
}

/* !
\param statusCode : RTSP status code
\return desc :RTSP status desc
\return ok :true for ok,false for not default status
*/
func GetRTSPStatusDesc(statusCode int) (desc string, ok bool) {
	desc, ok = statusCodeDescMap[statusCode]
	return
}
