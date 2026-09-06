package io.steviemul.orthanc.event.collector.service;

import io.steviemul.orthanc.event.Event;
import io.steviemul.orthanc.event.collector.config.KafkaConfig;
import lombok.RequiredArgsConstructor;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class EventPublisher {

  private final KafkaTemplate<String, Event> kafkaTemplate;
  private final KafkaConfig kafkaConfig;

  public void publishEvent(Event event) {
    kafkaTemplate.send(kafkaConfig.getEventTopic(), event);
  }
}
