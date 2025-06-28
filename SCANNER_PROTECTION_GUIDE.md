# 🛡️ Scanner Protection Guide for gitignore.lol

## Problem: Vulnerability Scanners Overwhelming Your Raspberry Pi

Vulnerability scanners constantly probe web services looking for security holes. They make hundreds of requests to common paths like `/wp-admin`, `/phpmyadmin`, `/config.php`, etc. This can easily overwhelm a Raspberry Pi, causing legitimate requests to fail with 404s.

## 🚀 Solution: Enhanced Rate Limiting with Scanner Protection

I've implemented an enhanced rate limiter specifically designed to handle scanner attacks on resource-constrained devices like your Raspberry Pi.

## ⚡ Quick Start - Enable Scanner Protection

```bash
# Enable enhanced rate limiting with aggressive scanner protection
./gitignore-server --enhanced-limiter \
                   --rate-limit 50 \
                   --error-rate-limit 5 \
                   --rate-window 30 \
                   --block-minutes 10 \
                   --max-violations 2

# For heavy attack scenarios (more aggressive)
./gitignore-server --enhanced-limiter \
                   --rate-limit 20 \
                   --error-rate-limit 3 \
                   --rate-window 60 \
                   --block-minutes 30 \
                   --max-violations 1
```

## 🔧 Configuration Options

### Enhanced Rate Limiter Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--enhanced-limiter` | `false` | Enable enhanced rate limiter with scanner protection |
| `--error-rate-limit` | `10` | Max 404/error requests per window per IP |
| `--block-minutes` | `5` | Minutes to block IPs that exceed limits |
| `--max-violations` | `3` | Max violations before exponential blocking |

### Standard Rate Limiter Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--rate-limit` | `100` | Max normal requests per window per IP |
| `--rate-window` | `60` | Rate limiting window in seconds |
| `--rate-cleanup` | `30000` | Cleanup interval in milliseconds |

## 🎯 How It Works

### 1. **Immediate Scanner Path Blocking**
Automatically blocks requests to common vulnerability scanner paths:
- WordPress: `/wp-admin`, `/wp-login`, `/xmlrpc.php`
- Admin panels: `/admin`, `/phpmyadmin`, `/administrator`  
- Config files: `/.env`, `/config`, `/web.config`
- Development: `/vendor`, `/node_modules`, `/.git`
- And 40+ more common scanner targets

### 2. **Dual Rate Limiting**
- **Normal requests**: Standard API usage (higher limits)
- **Error requests**: 404s and errors (much lower limits)

### 3. **Progressive Blocking**
- First violations: Short-term blocks (5-10 minutes)
- Repeat offenders: Exponential blocking (hours)
- Scanner paths: Immediate permanent blocks

### 4. **Memory Efficient**
- Automatic cleanup of expired entries
- Gradual violation forgiveness for reformed IPs
- Optimized for Raspberry Pi resource constraints

## 📊 Monitoring

### Check Protection Status
```bash
curl http://localhost:4444/stats
```

**Enhanced Stats Response:**
```json
{
  "enhanced_rate_limiting": "enabled",
  "stats": {
    "active_ips": 12,
    "blocked_ips": 3,
    "normal_requests": 89,
    "error_requests": 15,
    "normal_limit": 50,
    "error_limit": 5,
    "window_seconds": 30,
    "block_minutes": 10
  },
  "scanner_protection": "active"
}
```

### Key Metrics to Watch
- **blocked_ips**: Number of currently blocked IPs
- **error_requests**: High numbers indicate scanner activity
- **active_ips**: Total IPs being tracked

## ⚙️ Recommended Configurations

### For Light Scanner Activity
```bash
./gitignore-server --enhanced-limiter \
                   --rate-limit 100 \
                   --error-rate-limit 10 \
                   --block-minutes 5
```

### For Moderate Scanner Activity  
```bash
./gitignore-server --enhanced-limiter \
                   --rate-limit 50 \
                   --error-rate-limit 5 \
                   --block-minutes 15 \
                   --max-violations 2
```

### For Heavy Scanner Attacks
```bash
./gitignore-server --enhanced-limiter \
                   --rate-limit 20 \
                   --error-rate-limit 2 \
                   --rate-window 60 \
                   --block-minutes 60 \
                   --max-violations 1
```

### For Public APIs (Balanced)
```bash
./gitignore-server --enhanced-limiter \
                   --rate-limit 30 \
                   --error-rate-limit 3 \
                   --rate-window 30 \
                   --block-minutes 10
```

## 🔍 What Gets Blocked

### Scanner Paths (Immediate Block)
- WordPress: `/wp-*`, `/xmlrpc.php`
- Admin panels: `/admin*`, `/phpmyadmin`
- Config files: `/.env`, `/config*`, `/.git`
- Archives: `*.zip`, `*.tar`, `*.rar`
- Scripts: `*.php`, `*.asp`, `*.jsp`, `*.cgi`
- Logs: `*.log`, `*.txt`, `*.conf`

### Scanner Behaviors
- Too many 404 requests (exceeds `error-rate-limit`)
- Requests to non-existent file extensions
- Probing development directories

## 🚨 What Stays Available

- **Your legitimate API**: `/api/list`, `/api/go,node` etc.
- **Documentation**: `/swagger`, `/documentation`
- **Static files**: CSS, JS, images
- **Stats endpoint**: `/stats` for monitoring

## 💡 Additional Protection Tips

### 1. **Use a Reverse Proxy**
Consider putting Nginx or Cloudflare in front for additional protection:

```nginx
# nginx.conf snippet
location / {
    limit_req zone=general burst=10 nodelay;
    proxy_pass http://localhost:4444;
}

location ~ \.(php|asp|aspx|jsp)$ {
    return 404;
}
```

### 2. **Fail2Ban Integration**
Monitor your logs and ban repeat offenders at the firewall level:

```bash
# Example fail2ban rule
[gitignore-scanner]
enabled = true
filter = gitignore-scanner
logpath = /var/log/gitignore.log
maxretry = 3
bantime = 3600
```

### 3. **CloudFlare Protection**
If using CloudFlare, enable:
- Bot Fight Mode
- Rate Limiting rules
- Security Level: High

## 🔄 Migration from Basic Rate Limiter

If you're currently using the basic rate limiter, migration is seamless:

```bash
# Old way
./gitignore-server --rate-limit 100

# New way (enhanced protection)
./gitignore-server --enhanced-limiter --rate-limit 100 --error-rate-limit 5
```

Your existing configuration options will continue to work!

## 🎯 Expected Results

After enabling enhanced rate limiting:

- ✅ **Reduced 404s**: Scanner requests blocked before processing
- ✅ **Lower CPU usage**: Less work processing invalid requests  
- ✅ **Better responsiveness**: Legitimate requests get through faster
- ✅ **Automatic protection**: No manual intervention needed
- ✅ **Memory efficient**: Designed for Raspberry Pi constraints

Your Raspberry Pi should now handle scanner attacks gracefully while maintaining excellent performance for legitimate users! 🚀 