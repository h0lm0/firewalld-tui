# 🔥 firewalld-tui

Simple Go binary to better manage a given zone of firewalld


## 📦 Install

```bash
git clone https://github.com/h0lm0/firewalld-tui.git
cd firewalld-tui
make build
sudo make install
```

## 💡 Usage

```bash
# this command use the default zone 'restricted'
sudo firewalld-tui

# if you need to manage a special zone:
sudo firewalld-tui --zone=my-specific-zone
```
