# --- Stage 1: Build the Spring Boot Application ---
FROM eclipse-temurin:25-jdk AS builder

WORKDIR /app

# Copy Maven wrapper and reactor/module poms first for dependency layer caching
COPY .mvn/ .mvn/
COPY mvnw pom.xml ./
COPY common/pom.xml common/pom.xml
COPY event-collector/pom.xml event-collector/pom.xml
RUN ./mvnw -B -pl event-collector -am dependency:go-offline

# Copy source code and build
COPY common/src common/src
COPY event-collector/src event-collector/src
RUN ./mvnw -B -pl event-collector -am -DskipTests package

# --- Stage 2: Final Runtime Image ---
FROM eclipse-temurin:25-jre

WORKDIR /app

COPY --from=builder /app/event-collector/target/event-collector-*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
