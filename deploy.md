# Deployment Guide: Mtracker on GCP (Google Cloud Platform)

This guide provides step-by-step instructions to deploy the Mtracker backend and compile the Android app for production using Google Cloud Platform (GCP).

---

## Part 1: Deploying the Backend (Go API + Postgres)

We will deploy the backend to a **Google Compute Engine (GCE) VM instance** running Docker Compose.

### Step 1: Create a GCP VM Instance (Free Tier Eligible)
1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Navigate to **Compute Engine > VM Instances**.
3. Click **Create Instance**.
4. Configure the VM:
   - **Name:** `mtracker-backend`
   - **Region:** Choose one of the following (always free tier regions): `us-central1` (Iowa), `us-east1` (South Carolina), or `us-west1` (Oregon).
   - **Machine family:** General-purpose.
   - **Series:** `E2`.
   - **Machine type:** `e2-micro` (2 vCPU, 1 GB memory - Free-tier eligible).
   - **Boot disk:** Click **Change**:
     - **Operating System:** `Ubuntu`
     - **Version:** `Ubuntu 22.04 LTS` (or latest LTS)
     - **Boot disk type:** `Standard persistent disk` (Free tier covers up to 30 GB total).
     - **Size:** `15 GB` (sufficient for OS, Postgres, and Go API).
   - **Firewall:** Check both **Allow HTTP traffic** and **Allow HTTPS traffic**.
5. Click **Create**.

### Step 2: Configure GCP Firewall for Port `8080`
By default, the Go API runs on port `8080`. We need to open this port to allow the mobile app to communicate with it.
1. In the GCP Console, search for **VPC network > Firewall**.
2. Click **Create Firewall Rule**.
3. Configure the rule:
   - **Name:** `allow-mtracker-api`
   - **Network:** `default`
   - **Targets:** `All instances in the network` (or specify target tags if preferred)
   - **Source IPv4 ranges:** `0.0.0.0/0` (allows access from anywhere)
   - **Protocols and ports:** Check **Specified protocols and ports**, check **TCP**, and enter `8080`.
4. Click **Create**.

### Step 3: Install Docker and Docker Compose on the VM
Once the VM instance status is "Running", click the **SSH** button next to it in the Console.
Run the following commands in the SSH terminal to install Docker and Docker Compose:

```bash
# Update system package list
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io

# Start and enable Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Allow your user to run Docker without sudo (optional, requires logging back in)
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### Step 4: Clone Repository and Start Services
In the VM's SSH terminal:

1. Clone your project repository:
   ```bash
   git clone <YOUR_MTRACKER_GIT_REPOSITORY_URL> mtracker
   cd mtracker
   ```
2. Copy the example environment file to create production configuration:
   ```bash
   cp .env.example .env
   ```
3. Open `.env` and configure production values:
   ```bash
   nano .env
   ```
   - Update `POSTGRES_PASSWORD` to a secure custom password.
   - Update `JWT_SECRET` to a random long string for security.
   - Leave `PORT=8080` and `POSTGRES_HOST=localhost` (Docker Compose overrides this appropriately).
4. Launch the services:
   ```bash
   sudo docker-compose up -d
   ```
5. Verify the backend is up and running by visiting the health check in your local machine's web browser:
   ```
   http://<YOUR_VM_PUBLIC_IP>:8080/health
   ```

---

## Part 2: Building the Android Mobile App for Production

To build the release-ready standalone Android app (`.apk`) pointing to your live GCP backend:

### Step 1: Configure Production API URL
On your local computer, open the file `apps/mobile/.env` and update the backend URL to your VM's public IP:

```env
EXPO_PUBLIC_API_URL=http://<YOUR_VM_PUBLIC_IP>:8080
EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=your_android_oauth_client_id
EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=your_web_oauth_client_id
```

### Step 2: Generate Android Project Files
Expo projects require a "prebuild" step to generate the native Android code files before Gradle compilation.
Run the following in your local terminal:

```bash
cd apps/mobile
npx expo prebuild --platform android
```
*(This command will generate the `apps/mobile/android/` directory containing all native Android project resources.)*

### Step 3: Build the Standalone Release APK
Build the project using Gradle. Run this from `apps/mobile/android/`:

```bash
cd android
./gradlew assembleRelease
```
Once the build completes successfully, you will find your installable production APK file here:
`apps/mobile/android/app/build/outputs/apk/release/app-release.apk`

---

## Part 3: Hosting the APK on GCP Cloud Storage (Optional)

To make it easy to distribute the app to yourself or testers:

1. Go to the GCP Console and navigate to **Cloud Storage > Buckets**.
2. Click **Create**.
3. Set a unique bucket name (e.g. `mtracker-app-dist`) and click **Create**.
4. Upload your generated `app-release.apk` file.
5. Grant public read access to the file:
   - Click the **Permissions** tab of the uploaded file/bucket.
   - Click **Grant Access**.
   - Add Principal: `allUsers`
   - Assign Role: `Storage Object Viewer` (this makes the file publicly downloadable).
6. Copy the **Public URL** of the object to share or download the APK onto physical devices.
