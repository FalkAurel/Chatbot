from requests import get, RequestException, Response
from data import User
import streamlit as st

LOGIN: str = "http://backend:8080/api/login"

def login():
    """
    Login screen. 

    Is used to access the chatbot. This function may fail due to two reasons: 
    
    1. The credentials were rejected by the backend
    2. The database in the backend had a fatal error

    The data send to the Backend is a JSON-object

    ```
    {
        "email": str,
        "password": str,
    }
    ```

    And in the success case it'll return a JSON-object:

    ```
    {
        "Username": str,
        "JWTToken": str,
        "IsPremium": bool,
        "IsAdmin": bool
    }
    ```

    This JSON-object will be used to create `User` object
    which will be stored in the session_state under the key `user`.
    """
    st.set_page_config(
        page_title="Login", 
        page_icon="./Logo.png", 
        layout=None, 
        initial_sidebar_state=None, 
        menu_items=None
    )
    st.button("Back", on_click=lambda: st.session_state.update({"auth_page": "main"}), key="back_to_main")
    
    with st.form("Login"):
        email: str = st.text_input(label="email")
        password: str = st.text_input(label="Password", type="password")

        bool_email: bool = email != ""  # Simplified check
        bool_password: bool = password != ""
        
        if st.form_submit_button(label="Login"):
            try:
                response: Response = get(url=LOGIN, json={"email": email, "password": password})
                
                if response.status_code != 200:
                    st.error(f"Login failed: {response.content.decode('utf-8')}")
                else:
                    payload: dict = response.json()
                    st.session_state.user = User(
                        username=payload["Username"],
                        jwt=payload["JWTToken"],
                        is_premium=payload["IsPremium"],
                        is_admin=payload["IsAdmin"]
                    )
                    st.session_state.auth_page = 'main'
                    st.rerun()
            except RequestException as e:
                st.error(f"Login failed: {str(e)}")