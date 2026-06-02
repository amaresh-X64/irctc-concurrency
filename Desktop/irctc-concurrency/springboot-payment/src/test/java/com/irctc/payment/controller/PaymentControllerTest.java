package com.irctc.payment.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.irctc.constants.AppConstants;
import com.irctc.payment.dto.PaymentRequest;
import com.irctc.payment.dto.PaymentResponse;
import com.irctc.payment.service.PaymentService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.FilterType;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDateTime;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(controllers = PaymentController.class, excludeFilters = @ComponentScan.Filter(type = FilterType.ASSIGNABLE_TYPE, classes = com.irctc.middleware.JwtAuthFilter.class))
class PaymentControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private PaymentService paymentService;

    private PaymentRequest makeRequest() {
        PaymentRequest req = new PaymentRequest();
        req.setBookingId(1);
        req.setUserId(10);
        req.setAmount(1500.0);
        req.setPaymentMethod(AppConstants.METHOD_MOCK_UPI);
        return req;
    }

    private PaymentResponse makeResponse() {
        PaymentResponse resp = new PaymentResponse();
        resp.setPaymentId(42);
        resp.setBookingId(1);
        resp.setUserId(10);
        resp.setAmount(1500.0);
        resp.setStatus(AppConstants.PAYMENT_SUCCESS);
        resp.setTransactionId("TXN123ABC456");
        resp.setPnrNumber("PNR12345678");
        resp.setPaidAt(LocalDateTime.now());
        return resp;
    }

    @Test
    void processPayment_Returns200_WhenPaymentSucceeds() throws Exception {
        when(paymentService.processPayment(any())).thenReturn(makeResponse());

        mockMvc.perform(post("/api/v1/payments")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(makeRequest())))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value(AppConstants.MSG_PAYMENT_SUCCESS))
                .andExpect(jsonPath("$.data.paymentId").value(42))
                .andExpect(jsonPath("$.data.status").value(AppConstants.PAYMENT_SUCCESS))
                .andExpect(jsonPath("$.data.pnrNumber").value("PNR12345678"));
    }

    @Test
    void processPayment_Returns200_WithFailedStatus_WhenPaymentFails() throws Exception {
        PaymentResponse failedResp = makeResponse();
        failedResp.setStatus(AppConstants.PAYMENT_FAILED);
        failedResp.setPnrNumber(null);
        when(paymentService.processPayment(any())).thenReturn(failedResp);

        mockMvc.perform(post("/api/v1/payments")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(makeRequest())))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.status").value(AppConstants.PAYMENT_FAILED));
    }

    @Test
    void processPayment_Returns500_WhenServiceThrowsException() throws Exception {
        when(paymentService.processPayment(any())).thenThrow(new RuntimeException("DB down"));

        mockMvc.perform(post("/api/v1/payments")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(makeRequest())))
                .andExpect(status().isInternalServerError())
                .andExpect(jsonPath("$.success").value(false))
                .andExpect(jsonPath("$.message").value(AppConstants.MSG_SERVER_ERROR));
    }

    @Test
    void processPayment_IncludesTransactionId_InResponse() throws Exception {
        when(paymentService.processPayment(any())).thenReturn(makeResponse());

        mockMvc.perform(post("/api/v1/payments")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(makeRequest())))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.transactionId").value("TXN123ABC456"));
    }

    @Test
    void processPayment_ReturnsCorrectBookingAndUserIds() throws Exception {
        when(paymentService.processPayment(any())).thenReturn(makeResponse());

        mockMvc.perform(post("/api/v1/payments")
                .contentType(MediaType.APPLICATION_JSON)
                .content(objectMapper.writeValueAsString(makeRequest())))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.bookingId").value(1))
                .andExpect(jsonPath("$.data.userId").value(10));
    }

    @Test
    void getPaymentByBooking_Returns200_WhenPaymentExists() throws Exception {
        com.irctc.payment.model.Payment payment = new com.irctc.payment.model.Payment();
        payment.setId(42);
        payment.setBookingId(1);
        payment.setStatus(AppConstants.PAYMENT_SUCCESS);
        when(paymentService.getPaymentByBookingId(1)).thenReturn(payment);

        mockMvc.perform(get("/api/v1/payments/booking/1"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.data.bookingId").value(1))
                .andExpect(jsonPath("$.data.status").value(AppConstants.PAYMENT_SUCCESS));
    }

    @Test
    void getPaymentByBooking_Returns404_WhenPaymentNotFound() throws Exception {
        when(paymentService.getPaymentByBookingId(999))
                .thenThrow(new RuntimeException("Payment not found"));

        mockMvc.perform(get("/api/v1/payments/booking/999"))
                .andExpect(status().isNotFound());
    }

    @Test
    void getPaymentByBooking_Returns200_WithCorrectMessage() throws Exception {
        com.irctc.payment.model.Payment payment = new com.irctc.payment.model.Payment();
        payment.setId(1);
        payment.setBookingId(5);
        when(paymentService.getPaymentByBookingId(5)).thenReturn(payment);

        mockMvc.perform(get("/api/v1/payments/booking/5"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.message").value("Payment found"));
    }
}