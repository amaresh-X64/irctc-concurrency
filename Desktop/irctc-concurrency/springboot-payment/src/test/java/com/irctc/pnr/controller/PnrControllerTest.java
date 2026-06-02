package com.irctc.pnr.controller;

import com.irctc.constants.AppConstants;
import com.irctc.pnr.model.Pnr;
import com.irctc.pnr.service.PnrService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.FilterType;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDate;
import java.time.LocalDateTime;

import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(controllers = PnrController.class, excludeFilters = @ComponentScan.Filter(type = FilterType.ASSIGNABLE_TYPE, classes = com.irctc.middleware.JwtAuthFilter.class))
class PnrControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private PnrService pnrService;

    private Pnr makePnr() {
        Pnr pnr = new Pnr();
        pnr.setId(1);
        pnr.setPnrNumber("PNR12345678");
        pnr.setBookingId(10);
        pnr.setPaymentId(5);
        pnr.setUserId(3);
        pnr.setStatus(AppConstants.PNR_CONFIRMED);
        pnr.setJourneyDate(LocalDate.of(2028, 12, 25));
        pnr.setGeneratedAt(LocalDateTime.now());
        return pnr;
    }

    @Test
    void getPnr_Returns200_WhenPnrExists() throws Exception {
        when(pnrService.getPnrByNumber("PNR12345678")).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/PNR12345678"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value(AppConstants.MSG_PNR_GENERATED))
                .andExpect(jsonPath("$.data.pnrNumber").value("PNR12345678"))
                .andExpect(jsonPath("$.data.status").value(AppConstants.PNR_CONFIRMED));
    }

    @Test
    void getPnr_Returns404_WhenPnrNotFound() throws Exception {
        when(pnrService.getPnrByNumber("PNRNOTFOUND"))
                .thenThrow(new RuntimeException(AppConstants.MSG_PNR_NOT_FOUND));

        mockMvc.perform(get("/api/v1/pnr/PNRNOTFOUND"))
                .andExpect(status().isNotFound());
    }

    @Test
    void getPnr_ReturnsCorrectBookingId() throws Exception {
        when(pnrService.getPnrByNumber("PNR12345678")).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/PNR12345678"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.bookingId").value(10));
    }

    @Test
    void getPnr_ReturnsCorrectUserId() throws Exception {
        when(pnrService.getPnrByNumber("PNR12345678")).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/PNR12345678"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.userId").value(3));
    }

    @Test
    void getPnrByBooking_Returns200_WhenPnrExists() throws Exception {
        when(pnrService.getPnrByBookingId(10)).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/booking/10"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.success").value(true))
                .andExpect(jsonPath("$.message").value(AppConstants.MSG_PNR_GENERATED))
                .andExpect(jsonPath("$.data.pnrNumber").value("PNR12345678"));
    }

    @Test
    void getPnrByBooking_Returns404_WhenBookingHasNoPnr() throws Exception {
        when(pnrService.getPnrByBookingId(999))
                .thenThrow(new RuntimeException(AppConstants.MSG_PNR_NOT_FOUND));

        mockMvc.perform(get("/api/v1/pnr/booking/999"))
                .andExpect(status().isNotFound());
    }

    @Test
    void getPnrByBooking_ReturnsCorrectStatus() throws Exception {
        when(pnrService.getPnrByBookingId(10)).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/booking/10"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.status").value(AppConstants.PNR_CONFIRMED));
    }

    @Test
    void getPnrByBooking_ReturnsCorrectJourneyDate() throws Exception {
        when(pnrService.getPnrByBookingId(10)).thenReturn(makePnr());

        mockMvc.perform(get("/api/v1/pnr/booking/10"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.data.journeyDate").value("2028-12-25"));
    }
}