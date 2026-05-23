package com.irctc.pnr.service;

import com.irctc.constants.AppConstants;
import com.irctc.pnr.model.Pnr;
import com.irctc.pnr.repository.PnrRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.Random;

@Service
public class PnrService {

    private static final Logger log = LoggerFactory.getLogger(PnrService.class);
    private final PnrRepository pnrRepository;

    public PnrService(PnrRepository pnrRepository) {
        this.pnrRepository = pnrRepository;
    }

    public String generatePnr(Integer bookingId, Integer paymentId, Integer userId) {
        String pnrNumber = "PNR" + generateRandomDigits(8);

        Pnr pnr = new Pnr();
        pnr.setPnrNumber(pnrNumber);
        pnr.setBookingId(bookingId);
        pnr.setPaymentId(paymentId);
        pnr.setUserId(userId);
        pnr.setStatus(AppConstants.PNR_CONFIRMED);

        pnrRepository.save(pnr);
        log.info("PNR created: {} for booking: {}", pnrNumber, bookingId);
        return pnrNumber;
    }

    public Pnr getPnrByNumber(String pnrNumber) {
        return pnrRepository.findByPnrNumber(pnrNumber).orElseThrow(() -> new RuntimeException(AppConstants.MSG_PNR_NOT_FOUND));
    }

    public Pnr getPnrByBookingId(Integer bookingId) {
        return pnrRepository.findByBookingId(bookingId).orElseThrow(() -> new RuntimeException(AppConstants.MSG_PNR_NOT_FOUND));
    }

    private String generateRandomDigits(int length) {
        Random random = new Random();
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < length; i++) {
            sb.append(random.nextInt(10));
        }
        return sb.toString();
    }
}