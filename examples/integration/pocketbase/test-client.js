/**
 * Simple test client for the PocketBase MCP integration
 * Run this in a browser or with Node.js to test the connection
 */

const PORT = 8080; // Change this if you're using a custom port
const BASE_URL = `http://localhost:${PORT}/api/mcp`;

// Echo test function
async function testEcho() {
  console.log('Testing echo tool...');
  
  try {
    const response = await fetch(`${BASE_URL}/run`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        tool: 'echo',
        args: {
          message: 'Hello from test client!'
        }
      }),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Echo response:', data);
    return data;
  } catch (error) {
    console.error('Echo test failed:', error);
    throw error;
  }
}

// Weather test function
async function testWeather() {
  console.log('Testing weather tool...');
  
  try {
    const response = await fetch(`${BASE_URL}/run`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        tool: 'weather',
        args: {
          location: 'San Francisco'
        }
      }),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Weather response:', data);
    return data;
  } catch (error) {
    console.error('Weather test failed:', error);
    throw error;
  }
}

// Test SSE connection
function testSSE() {
  console.log('Testing SSE connection...');
  
  return new Promise((resolve, reject) => {
    const eventSource = new EventSource(`${BASE_URL}/sse`);
    
    eventSource.onopen = () => {
      console.log('SSE connection established');
    };
    
    eventSource.onmessage = (event) => {
      console.log('SSE message received:', event.data);
      eventSource.close();
      resolve(event.data);
    };
    
    eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      eventSource.close();
      reject(error);
    };
    
    // Close the connection after 5 seconds if no messages received
    setTimeout(() => {
      console.log('SSE test timeout - closing connection');
      eventSource.close();
      resolve('No messages received, but connection was established');
    }, 5000);
  });
}

// Run all tests
async function runTests() {
  try {
    console.log(`Testing connection to ${BASE_URL}`);
    
    // Test basic connectivity first
    const response = await fetch(`${BASE_URL}/tools`);
    if (!response.ok) {
      throw new Error(`Failed to connect to server: ${response.status}`);
    }
    const tools = await response.json();
    console.log('Available tools:', tools);
    
    // Run the tests
    await testEcho();
    await testWeather();
    await testSSE();
    
    console.log('All tests completed successfully!');
  } catch (error) {
    console.error('Tests failed:', error);
  }
}

// Run the tests when executed
runTests();

// Export for use in other contexts
if (typeof module !== 'undefined') {
  module.exports = {
    testEcho,
    testWeather,
    testSSE,
    runTests
  };
} 