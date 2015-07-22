package com.coduno;

import org.junit.Test;
import static org.junit.Assert.*;

public class TestApplication {
	@Test
	public void testOk() {
		Application app = new Application();
		assertEquals("ok", app.ok());
	}
}
