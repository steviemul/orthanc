package io.steviemul.orthanc.event.collector.controller;

import io.steviemul.orthanc.event.Event;
import io.steviemul.orthanc.event.collector.service.EventPublisher;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@Slf4j
@RestController
@RequestMapping("/events")
@RequiredArgsConstructor
public class EventController {

  private final EventPublisher eventPublisher;

  @GetMapping
  public String status() {
    return "OK";
  }

  @PostMapping
  public ResponseEntity<Void> processEvent(@RequestBody Event event) {
    log.info("Received event [{}]", event);

    eventPublisher.publishEvent(event);

    return ResponseEntity.noContent().build();
  }
}
