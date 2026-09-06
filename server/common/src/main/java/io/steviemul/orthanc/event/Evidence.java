package io.steviemul.orthanc.event;

import lombok.Builder;
import lombok.Getter;

import java.util.Map;

@Builder
@Getter
public class Evidence {

  private Map<String, String> facts;
}
