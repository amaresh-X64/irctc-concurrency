package com.irctc.pnr.controller;

import com.irctc.constants.AppConstants;
import com.irctc.helpers.ApiResponse;
import com.irctc.pnr.model.Pnr;
import com.irctc.pnr.service.PnrService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/pnr")
@CrossOrigin(origins = "*")
public class PnrController {

    private final PnrService pnrService;

    public PnrController(PnrService pnrService) {
        this.pnrService = pnrService;
    }

    @GetMapping("/{pnrNumber}")
    public ResponseEntity<ApiResponse<Pnr>> getPnr(@PathVariable String pnrNumber) {
        try {
            Pnr pnr = pnrService.getPnrByNumber(pnrNumber);
            return ResponseEntity.ok(ApiResponse.success(AppConstants.MSG_PNR_GENERATED, pnr));
        } catch (Exception e) {
            return ResponseEntity.notFound().build();
        }
    }

    @GetMapping("/booking/{bookingId}")
    public ResponseEntity<ApiResponse<Pnr>> getPnrByBooking(@PathVariable Integer bookingId) {
        try {
            Pnr pnr = pnrService.getPnrByBookingId(bookingId);
            return ResponseEntity.ok( ApiResponse.success(AppConstants.MSG_PNR_GENERATED, pnr)
            );
        } catch (Exception e) {
            return ResponseEntity.notFound().build();
        }
    }
}