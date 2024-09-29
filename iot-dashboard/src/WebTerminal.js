// // import React, { useRef, useEffect } from 'react'
// // import { useXTerm } from 'react-xtermjs'

// // const WebTerminal = () => {
// //   const { instance, ref } = useXTerm()
// //   instance?.writeln('Hello from react-xtermjs!')
// //   instance?.onData((data) => {console.log(data);instance?.write(data)})

// //   return <div ref={ref} style={{ width: '100%', height: '100%' }} />
// // }

// // export default WebTerminal;

// import React, { useRef, useEffect, useState } from 'react'
// import { useXTerm } from 'react-xtermjs'

// const WebTerminal = () => {
//   const commandRef = useRef('') // Initialize an empty command ref
//   // websocket
//   const [ws, setWs] = useState(null);
//   //const [message, setMessage] = useState('');
//   const message = useRef('')

//   useEffect(() => {
//     const wsUrl = 'ws://localhost:8080/websocket';
//     const wsOptions = {
//       // You can add headers or other options here
//     };

//     //const ws = new WebSocket(wsUrl, wsOptions);
//     const ws = new WebSocket(wsUrl);

//     ws.onopen = () => {
//       console.log('Connected to the WebSocket server');
//     };

//     ws.onmessage = (event) => {
//       console.log(`Received message from the server: ${event.data}`);
//       //setMessage(event.data);
//       message.current = event.data
//       console.log(`Received message from the server: ${message.current}`);
//     };

//     ws.onerror = (event) => {
//       console.log(`Error occurred: ${event}`);
//     };

//     ws.onclose = () => {
//       console.log('Disconnected from the WebSocket server');
//     };

//     setWs(ws);

//     return () => {
//       ws.close();
//     };
//   }, []);

//   const sendMessage = (message) => {
//     if (ws) {
//       ws.send(message);
//       return message.current
//     }
//   };

//   const { instance, ref } = useXTerm({
//     terminal: {
//       bellStyle: 'none',
//       cursorBlink: true,
//       fontSize: 12,
//       fontFamily: 'monospace',
//       theme: {
//         background: '#000',
//         foreground: '#fff',
//       },
//     },
//   })

//   instance?.write('Hello from react-xtermjs!\r\n$ ')
//   instance?.onData((data) => {
//     if (data === '\x7f' || data === '\b') { // Backspace character
//       commandRef.current = commandRef.current.slice(0, -1) // Update command ref by removing last character
//       instance?.write('\b \b') // Move cursor back and overwrite with space
//     } else if (data !== '\r' && data !== '\n') {
//       commandRef.current += data // Update command ref by appending new character
//       console.log(data)
//       instance?.write(data)
//     } else if (data === '\r' || data === '\n') {
//       // You can process the complete command here
//       console.log(`Complete command: ${commandRef.current}`)
//       sendMessage(commandRef.current)
//       instance?.write(message.current)
//       instance?.write("\r\n$ ")
//       commandRef.current = '' // Reset command ref
      
//     }
//   })
  

//   return <div ref={ref} style={{ width: '100%', height: '100%' }} />
// }

// export default WebTerminal

import React, { useRef, useEffect, useState } from 'react'
import { useXTerm } from 'react-xtermjs'

const WebTerminal = () => {
  const commandRef = useRef('') // Initialize an empty command ref
  // websocket
  const [ws, setWs] = useState(null);
  const messageRef = useRef('')
  const instanceRef = useRef(null)

  useEffect(() => {
    const wsUrl = 'ws://localhost:8080/websocket';
    const wsOptions = {
      // You can add headers or other options here
    };

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('Connected to the WebSocket server');
    };

    ws.onmessage = (event) => {
      console.log(`Received message from the server: ${event.data}`);
      messageRef.current = event.data
      console.log(`Received message from the server: ${messageRef.current.toString()}`);
      instanceRef.current?.write(messageRef.current.replaceAll("\n", "\r\n").toString())
      instanceRef.current?.write("$ ")
    };

    ws.onerror = (event) => {
      console.log(`Error occurred: ${event}`);
    };

    ws.onclose = () => {
      console.log('Disconnected from the WebSocket server');
    };

    setWs(ws);

    return () => {
      ws.close();
    };
  }, []);

  const sendMessage = (message) => {
    if (ws) {
      ws.send(message);
    }
  };

  const { instance, ref } = useXTerm({
    terminal: {
      bellStyle: 'none',
      cursorBlink: true,
      fontSize: 12,
      fontFamily: 'monospace',
      theme: {
        background: '#000',
        foreground: '#fff',
      },
    },
  })

  instanceRef.current = instance

  instance?.write('Hello from react-xtermjs!\r\n$ ')
  instance?.onData((data) => {
    if (data === '\x7f' || data === '\b') { // Backspace character
      commandRef.current = commandRef.current.slice(0, -1) // Update command ref by removing last character
      instance?.write('\b \b') // Move cursor back and overwrite with space
    } else if (data !== '\r' && data !== '\n') {
      commandRef.current += data // Update command ref by appending new character
      console.log(data)
      instance?.write(data)
    } else if (data === '\r' || data === '\n') {
      instance?.write("\r\n")
      // You can process the complete command here
      console.log(`Complete command: ${commandRef.current}`)
      sendMessage(commandRef.current)
      //instance?.write("$ ")
      commandRef.current = '' // Reset command ref
    }
  })
  

  return <div ref={ref} style={{ width: '100%', height: '100%' }} />
}

export default WebTerminal