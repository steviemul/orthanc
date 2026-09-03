package io.steviemul.orthanc.event.collector.controller;

import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import tools.jackson.databind.JsonNode;

@Slf4j
@RestController
@RequestMapping("/events")
public class EventController {

  @GetMapping
  public String status() {
    return "OK";
  }

  @PostMapping
  public ResponseEntity<Void> processEvent(@RequestBody JsonNode event) {
    log.info("Received event [{}]", event);

    return ResponseEntity.noContent().build();
  }
}
