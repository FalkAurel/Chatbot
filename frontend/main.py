import streamlit as st
from chatbot import chatbot
from login import login
from signup import sign_up
from admin_dashboard import admin_dashboard

def main():
    st.set_page_config(
        page_title="LawChina-AI", 
        page_icon="./Logo.png", 
        layout=None, 
        initial_sidebar_state=None, 
        menu_items=None
    )

    st.image("./assets/Logo-AI-LC-caiyun.svg")
    st.title("LC-MingBai")
    
    if 'auth_page' not in st.session_state:
        st.session_state.auth_page = 'main'  # 'main', 'signup', or 'login'
    
    if not st.session_state.get("user"):
        if st.session_state.auth_page == 'main':
            signup, login_in = st.columns([2, 2])
            with signup:
                if st.button("Sign up"):
                    st.session_state.auth_page = 'signup'
                    st.rerun()
            with login_in:
                if st.button("Login"):
                    st.session_state.auth_page = 'login'
                    st.rerun()
        elif st.session_state.auth_page == 'signup':
            sign_up()
        elif st.session_state.auth_page == 'login':
            login()
    else:
        if st.session_state.get("Admin Dashboard"):
            admin_dashboard(user=st.session_state.get("user"))
        else:
            chatbot()

if __name__ == "__main__":
    main()