package com.irctc.pnr.repository;

import com.irctc.pnr.model.Pnr;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface PnrRepository extends JpaRepository<Pnr, Integer> {

    Optional<Pnr> findByPnrNumber(String pnrNumber);
    Optional<Pnr> findByBookingId(Integer bookingId);
}