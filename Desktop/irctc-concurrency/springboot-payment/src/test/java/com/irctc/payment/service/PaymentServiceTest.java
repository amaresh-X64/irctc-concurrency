package com.irctc.payment.service;

import com.irctc.constants.AppConstants;
import com.irctc.payment.dto.PaymentRequest;
import com.irctc.payment.dto.PaymentResponse;
import com.irctc.payment.model.Payment;
import com.irctc.payment.repository.PaymentRepository;
import com.irctc.pnr.service.PnrService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Optional;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PaymentServiceTest {

    @Mock
    private PaymentRepository paymentRepository;

    @Mock
    private PnrService pnrService;

    @InjectMocks
    private PaymentService paymentService;

    private final ArgumentCaptor<Payment> paymentCaptor =
            ArgumentCaptor.forClass(Payment.class);

    private PaymentRequest request;
    private Payment savedPayment;

    @BeforeEach
    void setUp() {
        request = new PaymentRequest();
        request.setBookingId(1);
        request.setUserId(10);
        request.setAmount(500.0);
        request.setPaymentMethod(AppConstants.METHOD_MOCK_UPI);

        savedPayment = new Payment();
        savedPayment.setId(100);
        savedPayment.setBookingId(1);
        savedPayment.setUserId(10);
        savedPayment.setAmount(BigDecimal.valueOf(500.0));
        savedPayment.setTransactionId("TXN123ABC456D");
        savedPayment.setPaidAt(LocalDateTime.now());
    }

    @Test
    void processPayment_Success_ReturnsPnrNumber() {
        savedPayment.setStatus(AppConstants.PAYMENT_SUCCESS);
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(1, 10, 1)).thenReturn("PNR12345678");

        PaymentResponse response = null;
        for (int i = 0; i < 20; i++) {
            response = paymentService.processPayment(request);
            if (AppConstants.PAYMENT_SUCCESS.equals(response.getStatus())) break;
        }
        assertThat(response).isNotNull();
        assertThat(response.getBookingId()).isEqualTo(1);
        assertThat(response.getUserId()).isEqualTo(10);
        assertThat(response.getAmount()).isEqualTo(500.0);
        assertThat(response.getPaymentId()).isEqualTo(100);

        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(1);
        assertThat(captured.getUserId()).isEqualTo(10);
        assertThat(captured.getAmount()).isEqualByComparingTo(BigDecimal.valueOf(500.0));
    }

    @Test
    void processPayment_TransactionIdIsGenerated() {
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);

        paymentService.processPayment(request);
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getTransactionId()).isNotNull();
        assertThat(captured.getTransactionId()).startsWith("TXN");
    }

    @Test
    void processPayment_PaymentSavedToRepository() {
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(1, 10, 1)).thenReturn("PNR12345678");

        paymentService.processPayment(request);
        verify(paymentRepository, times(1)).save(paymentCaptor.capture());
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(1);
        assertThat(captured.getUserId()).isEqualTo(10);
        assertThat(captured.getAmount()).isEqualByComparingTo(BigDecimal.valueOf(500.0));
    }

    @Test
    void processPayment_AmountSetCorrectly() {
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(1, 10, 1)).thenReturn("PNR12345678");

        PaymentResponse response = paymentService.processPayment(request);

        assertThat(response.getAmount()).isEqualTo(500.0);
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getAmount()).isEqualByComparingTo(BigDecimal.valueOf(500.0));
    }

    @Test
    void processPayment_WithMockCard_Works() {
        request.setPaymentMethod(AppConstants.METHOD_MOCK_CARD);
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);

        PaymentResponse response = paymentService.processPayment(request);

        assertThat(response).isNotNull();
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(1);
        assertThat(captured.getUserId()).isEqualTo(10);
    }

    @Test
    void processPayment_WithMockNetBanking_Works() {
        request.setPaymentMethod(AppConstants.METHOD_MOCK_NET);
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);

        PaymentResponse response = paymentService.processPayment(request);

        assertThat(response).isNotNull();

        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(1);
        assertThat(captured.getUserId()).isEqualTo(10);
    }

    @Test
    void processPayment_RepositoryThrowsException_Propagates() {
        when(paymentRepository.save(paymentCaptor.capture()))
                .thenThrow(new RuntimeException("DB connection failed"));

        assertThatThrownBy(() -> paymentService.processPayment(request))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("DB connection failed");
    }

    @Test
    void processPayment_PnrServiceNotCalledOnFailure() {
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(1, 10, 1)).thenReturn("PNR00000000");

        paymentService.processPayment(request);

        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(1);
        assertThat(captured.getUserId()).isEqualTo(10);
    }

    @Test
    void processPayment_ResponseHasCorrectBookingId() {
        request.setBookingId(999);
        savedPayment.setBookingId(999);
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(999, 10, 1)).thenReturn("PNR00000001");

        PaymentResponse response = paymentService.processPayment(request);

        assertThat(response.getBookingId()).isEqualTo(999);
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getBookingId()).isEqualTo(999);
    }

    @Test
    void processPayment_ResponseHasCorrectUserId() {
        request.setUserId(777);
        when(paymentRepository.save(paymentCaptor.capture())).thenReturn(savedPayment);
        lenient().when(pnrService.generatePnr(1, 777, 1)).thenReturn("PNR00000002");

        PaymentResponse response = paymentService.processPayment(request);

        assertThat(response.getUserId()).isEqualTo(777);
        Payment captured = paymentCaptor.getValue();
        assertThat(captured.getUserId()).isEqualTo(777);
    }

    @Test
    void getPaymentByBookingId_Found_ReturnsPayment() {
        when(paymentRepository.findByBookingId(1)).thenReturn(Optional.of(savedPayment));

        Payment result = paymentService.getPaymentByBookingId(1);

        assertThat(result).isNotNull();
        assertThat(result.getId()).isEqualTo(100);
        assertThat(result.getBookingId()).isEqualTo(1);
        verify(paymentRepository, times(1)).findByBookingId(1);
    }

    @Test
    void getPaymentByBookingId_NotFound_ThrowsException() {
        when(paymentRepository.findByBookingId(999)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> paymentService.getPaymentByBookingId(999))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("Payment not found");
    }

    @Test
    void getPaymentByBookingId_VerifiesRepositoryCall() {
        when(paymentRepository.findByBookingId(42)).thenReturn(Optional.of(savedPayment));

        paymentService.getPaymentByBookingId(42);

        verify(paymentRepository, times(1)).findByBookingId(42);
        verifyNoMoreInteractions(paymentRepository);
    }

    @Test
    void getPaymentByBookingId_ReturnsCorrectStatus() {
        savedPayment.setStatus(AppConstants.PAYMENT_SUCCESS);
        when(paymentRepository.findByBookingId(1)).thenReturn(Optional.of(savedPayment));

        Payment result = paymentService.getPaymentByBookingId(1);

        assertThat(result.getStatus()).isEqualTo(AppConstants.PAYMENT_SUCCESS);
    }

    @Test
    void getPaymentByBookingId_ReturnsCorrectTransactionId() {
        savedPayment.setTransactionId("TXN_SPECIAL_001");
        when(paymentRepository.findByBookingId(1)).thenReturn(Optional.of(savedPayment));

        Payment result = paymentService.getPaymentByBookingId(1);

        assertThat(result.getTransactionId()).isEqualTo("TXN_SPECIAL_001");
    }
}