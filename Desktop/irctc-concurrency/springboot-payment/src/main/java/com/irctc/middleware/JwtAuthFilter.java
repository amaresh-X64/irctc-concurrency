package com.irctc.middleware;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

@Component
public class JwtAuthFilter extends OncePerRequestFilter {

    private final String jwtSecretKey;

    public JwtAuthFilter(@Value("${jwt.secret-key}") String jwtSecretKey) {
        if (jwtSecretKey == null || jwtSecretKey.isBlank()) {
            throw new IllegalStateException("jwt.secret-key must be set");
        }
        this.jwtSecretKey = jwtSecretKey;
    }

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        return request.getRequestURI().equals("/health")
                || request.getMethod().equals("OPTIONS");
    }

    @Override
    protected void doFilterInternal(
            HttpServletRequest request,
            HttpServletResponse response,
            FilterChain filterChain) throws ServletException, IOException {

        String authHeader = request.getHeader("Authorization");

        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            reject(response, "Please login to continue");
            return;
        }

        try {
            Claims claims = Jwts.parser()
                    .verifyWith(Keys.hmacShaKeyFor(
                            jwtSecretKey.getBytes(StandardCharsets.UTF_8)))
                    .build()
                    .parseSignedClaims(authHeader.substring(7))
                    .getPayload();

            request.setAttribute("user_id", claims.getSubject());
            filterChain.doFilter(request, response);

        } catch (JwtException e) {
            reject(response, "Session expired, please login again");
        }
    }

    private void reject(HttpServletResponse response, String message) throws IOException {
        response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
        response.setContentType(MediaType.APPLICATION_JSON_VALUE);
        response.getWriter().write(
                "{\"success\":false,\"message\":\"" + message + "\",\"data\":null}");
    }
}