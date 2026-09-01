package io.steviemul.orthanc.event.collector.controller;

import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@Slf4j
@RestController
public class EventController {

  @GetMapping
  public String status() {
    return "OK";
  }
}
