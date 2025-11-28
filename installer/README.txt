PeerTube Monitor - Post-Installation Setup
==========================================

The PeerTube Monitor service has been installed but requires configuration before it will start successfully.

CONFIGURATION STEPS:

1. Edit the configuration file:
   C:\Program Files\PeerTube Monitor\config.json

2. Update the following settings:
   - PeerTube server URL, username, and password
   - Watch, done, and failed folder paths
   - Video upload defaults (category, licence, language, privacy)

3. Create the folders specified in the configuration:
   - Watch folder (where you'll drop videos)
   - Done folder (for successful uploads)
   - Failed folder (for failed uploads)

4. Start the service:
   - Open Services (services.msc)
   - Find "PeerTube Monitor"
   - Right-click and select "Start"

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
