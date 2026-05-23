package com.irctc.payment.service;

import com.irctc.constants.AppConstants;
import com.irctc.payment.dto.PaymentRequest;
import com.irctc.payment.dto.PaymentResponse;
import com.irctc.payment.model.Payment;
import com.irctc.payment.repository.PaymentRepository;
import com.irctc.pnr.service.PnrService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;

@Service
public class PaymentService {

    private static final Logger log = LoggerFactory.getLogger(PaymentService.class);

    private final PaymentRepository paymentRepository;
    private final PnrService pnrService;

    public PaymentService(PaymentRepository paymentRepository, PnrService pnrService) {
        this.paymentRepository = paymentRepository;
        this.pnrService = pnrService;
    }

    @Transactional
    public PaymentResponse processPayment(PaymentRequest request) {
        log.info("Processing payment for booking: {}", request.getBookingId());

        String transactionId = "TXN" + UUID.randomUUID()
                .toString().replace("-", "").substring(0, 12).toUpperCase();

        boolean paymentSuccess = simulateMockPayment();

        Payment payment = new Payment();
        payment.setBookingId(request.getBookingId());
        payment.setUserId(request.getUserId());
        payment.setAmount(BigDecimal.valueOf(request.getAmount()));
        payment.setPaymentMethod(request.getPaymentMethod());
        payment.setTransactionId(transactionId);

        if (paymentSuccess) {
            payment.setStatus(AppConstants.PAYMENT_SUCCESS);
            payment.setPaidAt(LocalDateTime.now());
        } else {
            payment.setStatus(AppConstants.PAYMENT_FAILED);
        }

        Payment saved = paymentRepository.save(payment);
        log.info("Payment saved: {} - Status: {}", transactionId, payment.getStatus());

        String pnrNumber = null;
        if (paymentSuccess) {
            pnrNumber = pnrService.generatePnr(request.getBookingId(),saved.getId(),request.getUserId());
            log.info("PNR generated: {}", pnrNumber);
        }

        PaymentResponse response = new PaymentResponse();
        response.setPaymentId(saved.getId());
        response.setBookingId(request.getBookingId());
        response.setUserId(request.getUserId());
        response.setAmount(request.getAmount());
        response.setStatus(payment.getStatus());
        response.setTransactionId(transactionId);
        response.setPnrNumber(pnrNumber);
        response.setPaidAt(payment.getPaidAt());

        return response;
    }

    private boolean simulateMockPayment() {
        return Math.random() > 0.1;
    }

    public Payment getPaymentByBookingId(Integer bookingId) {
        return paymentRepository.findByBookingId(bookingId).orElseThrow(() -> new RuntimeException("Payment not found"));
    }
}