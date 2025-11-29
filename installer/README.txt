PeerTube Monitor – Post-Installation Setup
===========================================

REQUIRED: Configure before starting the service!

QUICK START:

1. Edit: C:\Program Files\PeerTube Monitor\config.json
   - Set your PeerTube server URL, username, and password
   - Set watch, done, and failed folder paths
   - Adjust video upload defaults (category, licence, language, privacy)

2. Create the folders you specified in config.json
   (The service will NOT create them automatically)

3. Start the service:
   - Press Win+R, type: services.msc
   - Find "PeerTube Monitor"
   - Right-click → Start

ALTERNATIVE: Use Environment Variables for Credentials

For better security, you can store credentials as service environment variables:

1. Open Registry Editor (regedit)
2. Navigate to: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\PeerTubeMonitor
3. Add a multi-string value named "Environment" with these entries:
   PEERTUBE_URL=https://your-peertube-instance.com
   PEERTUBE_USERNAME=your-username
   PEERTUBE_PASSWORD=your-password

4. Leave username and password empty in config.json

DOCUMENTATION:

Full documentation: https://github.com/dsu-teknik/peertube-monitor

TROUBLESHOOTING:

- Check logs: C:\Program Files\PeerTube Monitor\peertube-monitor.log
- Verify config.json syntax is valid JSON
- Ensure folders exist and are accessible
- Check PeerTube credentials are correct
