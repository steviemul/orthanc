# --- Stage 1: Build the Spring Boot Application ---
FROM eclipse-temurin:25-jdk AS builder

WORKDIR /app

# Copy Maven wrapper and pom first for dependency layer caching
COPY .mvn/ .mvn/
COPY mvnw pom.xml ./
RUN ./mvnw -B dependency:go-offline

# Copy source code and build
COPY src/ src/
RUN ./mvnw -B -DskipTests package

# --- Stage 2: Final Runtime Image ---
FROM eclipse-temurin:25-jre

WORKDIR /app

COPY --from=builder /app/target/event-collector-*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
