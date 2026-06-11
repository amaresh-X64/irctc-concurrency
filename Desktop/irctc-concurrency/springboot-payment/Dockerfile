FROM maven:3.9.6-eclipse-temurin-21 AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -q
COPY src ./src
RUN mvn package -DskipTests -q

# ── Runtime stage ─────────────────────────────────────────────────────────
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Download the OTel Java agent (single fat JAR, ~10MB)
# This agent instruments Spring Boot, Hibernate, JDBC, and HTTP clients
# automatically at JVM startup — no code changes needed.
ARG OTEL_AGENT_VERSION=2.3.0
RUN apk add --no-cache curl && \
    curl -fsSL \
      "https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v${OTEL_AGENT_VERSION}/opentelemetry-javaagent.jar" \
      -o /app/opentelemetry-javaagent.jar

COPY --from=builder /app/target/*.jar app.jar

# JAVA_TOOL_OPTIONS is picked up automatically by the JVM.
# The agent intercepts all Spring MVC requests, JPA queries, and
# outbound HTTP calls — each becomes a span sent to the OTel Collector.
ENV JAVA_TOOL_OPTIONS="-javaagent:/app/opentelemetry-javaagent.jar"

EXPOSE 8003
ENTRYPOINT ["java", "-jar", "app.jar"]
