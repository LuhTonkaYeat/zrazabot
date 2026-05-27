import random
import sqlite3
import os
os.environ['HTTP_PROXY'] = ''
os.environ['HTTPS_PROXY'] = ''
os.environ['ALL_PROXY'] = ''
from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
DB_PATH = os.path.join(BASE_DIR, 'zrazy.db')


def init_db():
    conn = sqlite3.connect(DB_PATH)
    cur = conn.cursor()
    cur.execute('''
        CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY,
            total INTEGER DEFAULT 0
        )
    ''')
    conn.commit()
    conn.close()


def add_zrazy(user_id: int, amount: int):
    conn = sqlite3.connect(DB_PATH)
    cur = conn.cursor()
    cur.execute('''
        INSERT INTO users (user_id, total) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET total = total + ?
    ''', (user_id, amount, amount))
    conn.commit()
    conn.close()


def get_total(user_id: int) -> int:
    conn = sqlite3.connect(DB_PATH)
    cur = conn.cursor()
    cur.execute('SELECT total FROM users WHERE user_id = ?', (user_id,))
    result = cur.fetchone()
    conn.close()
    return result[0] if result else 0


async def zraza_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    user = update.effective_user
    user_id = user.id
    eaten = random.randint(1, 10)

    add_zrazy(user_id, eaten)
    total = get_total(user_id)

    await update.message.reply_text(
        f"🍽 {user.first_name} съел {eaten} зразу(ы)!\n"
        f"📊 *Всего за историю:* {total} зраз.",
        parse_mode="Markdown"
    )


def main():
    init_db()
    TOKEN = "8693360984:AAHx4PYb38xQ1pnEVtKTWQcmR7s_vCp2kkA"

    app = Application.builder().token(TOKEN).build()
    app.add_handler(CommandHandler("zraza", zraza_command))

    print("Бот запущен! Напиши /zraza в Telegram")
    app.run_polling()


if __name__ == "__main__":
    main()