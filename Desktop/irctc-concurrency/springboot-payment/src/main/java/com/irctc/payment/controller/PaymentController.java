package com.irctc.payment.controller;

import com.irctc.constants.AppConstants;
import com.irctc.helpers.ApiResponse;
import com.irctc.payment.dto.PaymentRequest;
import com.irctc.payment.dto.PaymentResponse;
import com.irctc.payment.service.PaymentService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/payments")
@CrossOrigin(origins = "*")
public class PaymentController {

    private static final Logger log = LoggerFactory.getLogger(PaymentController.class);
    private final PaymentService paymentService;

    public PaymentController(PaymentService paymentService) {
        this.paymentService = paymentService;
    }

    @PostMapping
    public ResponseEntity<ApiResponse<PaymentResponse>> processPayment(
            @RequestBody PaymentRequest request) {
        try {
            PaymentResponse response = paymentService.processPayment(request);
            return ResponseEntity.ok(
                ApiResponse.success(AppConstants.MSG_PAYMENT_SUCCESS, response)
            );
        } catch (Exception e) {
            log.error("Payment error: {}", e.getMessage());
            return ResponseEntity.internalServerError().body(ApiResponse.error(AppConstants.MSG_SERVER_ERROR));
        }
    }

    @GetMapping("/booking/{bookingId}")
    public ResponseEntity<ApiResponse<?>> getPaymentByBooking(
            @PathVariable Integer bookingId) {
        try {
            var payment = paymentService.getPaymentByBookingId(bookingId);
            return ResponseEntity.ok(ApiResponse.success("Payment found", payment));
        } catch (Exception e) {
            return ResponseEntity.notFound().build();
        }
    }
}