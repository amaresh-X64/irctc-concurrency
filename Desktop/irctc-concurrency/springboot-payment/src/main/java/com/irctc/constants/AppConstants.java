package com.irctc.constants;

public class AppConstants {

    private AppConstants() {}

    public static final String PAYMENT_PENDING   = "PENDING";
    public static final String PAYMENT_SUCCESS   = "SUCCESS";
    public static final String PAYMENT_FAILED    = "FAILED";
    public static final String PAYMENT_REFUNDED  = "REFUNDED";

    public static final String PNR_CONFIRMED     = "CONFIRMED";
    public static final String PNR_CANCELLED     = "CANCELLED";

    public static final String METHOD_MOCK_UPI   = "MOCK_UPI";
    public static final String METHOD_MOCK_CARD  = "MOCK_CARD";
    public static final String METHOD_MOCK_NET   = "MOCK_NETBANKING";

    public static final String MSG_PAYMENT_SUCCESS  = "Payment successful";
    public static final String MSG_PAYMENT_FAILED   = "Payment failed";
    public static final String MSG_PNR_GENERATED    = "PNR generated successfully";
    public static final String MSG_PNR_NOT_FOUND    = "PNR not found";
    public static final String MSG_SERVER_ERROR     = "Internal server error";
}