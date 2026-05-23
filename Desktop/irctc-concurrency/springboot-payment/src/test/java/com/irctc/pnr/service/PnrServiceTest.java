package com.irctc.pnr.service;

import com.irctc.constants.AppConstants;
import com.irctc.pnr.model.Pnr;
import com.irctc.pnr.repository.PnrRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class PnrServiceTest {

    @Mock
    private PnrRepository pnrRepository;

    @InjectMocks
    private PnrService pnrService;
    private final ArgumentCaptor<Pnr> pnrCaptor =
            ArgumentCaptor.forClass(Pnr.class);

    private Pnr samplePnr;

    @BeforeEach
    void setUp() {
        samplePnr = new Pnr();
        samplePnr.setId(1);
        samplePnr.setPnrNumber("PNR12345678");
        samplePnr.setBookingId(10);
        samplePnr.setPaymentId(100);
        samplePnr.setUserId(5);
        samplePnr.setStatus(AppConstants.PNR_CONFIRMED);
    }

    @Test
    void generatePnr_ReturnsPnrNumber() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        String result = pnrService.generatePnr(10, 100, 5);

        assertThat(result).isNotNull();
        assertThat(result).startsWith("PNR");
        assertThat(result).hasSize(11);
    }

    @Test
    void generatePnr_SavesCorrectBookingId() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        pnrService.generatePnr(10, 100, 5);

        // use class-level captor — no need to declare inside test
        assertThat(pnrCaptor.getValue().getBookingId()).isEqualTo(10);
    }

    @Test
    void generatePnr_SavesCorrectPaymentId() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        pnrService.generatePnr(10, 100, 5);

        assertThat(pnrCaptor.getValue().getPaymentId()).isEqualTo(100);
    }

    @Test
    void generatePnr_SavesCorrectUserId() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        pnrService.generatePnr(10, 100, 5);

        assertThat(pnrCaptor.getValue().getUserId()).isEqualTo(5);
    }

    @Test
    void generatePnr_StatusIsConfirmed() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        pnrService.generatePnr(10, 100, 5);

        assertThat(pnrCaptor.getValue().getStatus()).isEqualTo(AppConstants.PNR_CONFIRMED);
    }

    @Test
    void generatePnr_SavesToRepository() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        pnrService.generatePnr(10, 100, 5);
        verify(pnrRepository, times(1)).save(pnrCaptor.capture());
    }

    @Test
    void generatePnr_EachCallGeneratesUniquePnr() {
        when(pnrRepository.save(pnrCaptor.capture())).thenReturn(samplePnr);

        String pnr1 = pnrService.generatePnr(1, 1, 1);
        String pnr2 = pnrService.generatePnr(2, 2, 2);

        assertThat(pnr1).isNotEqualTo(pnr2);
    }

    @Test
    void getPnrByNumber_Found_ReturnsPnr() {
        when(pnrRepository.findByPnrNumber("PNR12345678")).thenReturn(Optional.of(samplePnr));

        Pnr result = pnrService.getPnrByNumber("PNR12345678");

        assertThat(result).isNotNull();
        assertThat(result.getPnrNumber()).isEqualTo("PNR12345678");
        assertThat(result.getBookingId()).isEqualTo(10);
        verify(pnrRepository, times(1)).findByPnrNumber("PNR12345678");
    }

    @Test
    void getPnrByNumber_NotFound_ThrowsException() {
        when(pnrRepository.findByPnrNumber("PRNNOTFOUND")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> pnrService.getPnrByNumber("PRNNOTFOUND"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage(AppConstants.MSG_PNR_NOT_FOUND);
    }

    @Test
    void getPnrByNumber_ReturnsCorrectStatus() {
        when(pnrRepository.findByPnrNumber("PNR12345678")).thenReturn(Optional.of(samplePnr));

        Pnr result = pnrService.getPnrByNumber("PNR12345678");

        assertThat(result.getStatus()).isEqualTo(AppConstants.PNR_CONFIRMED);
    }

    @Test
    void getPnrByBookingId_Found_ReturnsPnr() {
        when(pnrRepository.findByBookingId(10)).thenReturn(Optional.of(samplePnr));

        Pnr result = pnrService.getPnrByBookingId(10);

        assertThat(result).isNotNull();
        assertThat(result.getBookingId()).isEqualTo(10);
        assertThat(result.getPnrNumber()).isEqualTo("PNR12345678");
        verify(pnrRepository, times(1)).findByBookingId(10);
    }

    @Test
    void getPnrByBookingId_NotFound_ThrowsException() {
        when(pnrRepository.findByBookingId(999)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> pnrService.getPnrByBookingId(999))
                .isInstanceOf(RuntimeException.class)
                .hasMessage(AppConstants.MSG_PNR_NOT_FOUND);
    }

    @Test
    void getPnrByBookingId_ReturnsCorrectPaymentId() {
        when(pnrRepository.findByBookingId(10)).thenReturn(Optional.of(samplePnr));

        Pnr result = pnrService.getPnrByBookingId(10);

        assertThat(result.getPaymentId()).isEqualTo(100);
    }

    @Test
    void getPnrByBookingId_ReturnsCorrectUserId() {
        when(pnrRepository.findByBookingId(10)).thenReturn(Optional.of(samplePnr));

        Pnr result = pnrService.getPnrByBookingId(10);

        assertThat(result.getUserId()).isEqualTo(5);
    }

    @Test
    void getPnrByBookingId_VerifiesRepositoryCall() {
        when(pnrRepository.findByBookingId(10)).thenReturn(Optional.of(samplePnr));

        pnrService.getPnrByBookingId(10);

        verify(pnrRepository, times(1)).findByBookingId(10);
        verifyNoMoreInteractions(pnrRepository);
    }
}